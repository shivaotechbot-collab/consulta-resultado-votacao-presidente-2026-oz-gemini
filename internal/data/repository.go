package data

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"consulta-resultado-votacao-presidente-2026-oz-gemini/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	db *sql.DB
}

type Filter struct {
	Zona      int
	Secao     int
	Candidate string
	Local     string
	Limit     int
	Offset    int
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS votacao (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nr_zona INTEGER,
		nm_local_votacao TEXT,
		ds_local_votacao_endereco TEXT,
		nr_secao INTEGER,
		nr_local_votacao INTEGER,
		ds_cargo TEXT,
		nr_votavel TEXT,
		nm_votavel TEXT,
		qt_aptos INTEGER,
		qt_comparecimento INTEGER,
		qt_abstencoes INTEGER,
		qt_votos_nominais INTEGER,
		qt_votos INTEGER
	);
	CREATE INDEX IF NOT EXISTS idx_zona_secao ON votacao(nr_zona, nr_secao);
	CREATE INDEX IF NOT EXISTS idx_votavel ON votacao(nr_votavel, nm_votavel);
	CREATE INDEX IF NOT EXISTS idx_local ON votacao(nm_local_votacao);
	`
	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("create tables: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

// decodeISO88591ToUTF8 handles ISO-8859-1 (Latin1) bytes converting to UTF-8 runes
func decodeISO88591ToUTF8(b []byte) string {
	runes := make([]rune, len(b))
	for i, c := range b {
		runes[i] = rune(c)
	}
	return string(runes)
}

func (r *Repository) ImportCSV(reader io.Reader) (int, error) {
	// Read entire stream or line-by-line while converting ISO-8859-1 to UTF-8
	bufReader := bufio.NewReader(reader)
	var utf8Buf bytes.Buffer

	for {
		line, err := bufReader.ReadBytes('\n')
		if len(line) > 0 {
			utf8Str := decodeISO88591ToUTF8(line)
			utf8Buf.WriteString(utf8Str)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, fmt.Errorf("read line: %w", err)
		}
	}

	csvReader := csv.NewReader(&utf8Buf)
	csvReader.Comma = ';'
	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err != nil {
		return 0, fmt.Errorf("read header: %w", err)
	}

	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[strings.TrimSpace(col)] = i
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO votacao (
			nr_zona, nm_local_votacao, ds_local_votacao_endereco,
			nr_secao, nr_local_votacao, ds_cargo, nr_votavel, nm_votavel,
			qt_aptos, qt_comparecimento, qt_abstencoes, qt_votos_nominais, qt_votos
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	parseInt := func(record []string, key string) int {
		idx, ok := colIdx[key]
		if !ok || idx >= len(record) {
			return 0
		}
		v, _ := strconv.Atoi(strings.TrimSpace(record[idx]))
		return v
	}

	parseStr := func(record []string, key string) string {
		idx, ok := colIdx[key]
		if !ok || idx >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[idx])
	}

	count := 0
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		zona := parseInt(rec, "nr_zona")
		secao := parseInt(rec, "nr_secao")
		nrLocal := parseInt(rec, "nr_local_votacao")
		local := parseStr(rec, "nm_local_votacao")
		end := parseStr(rec, "ds_local_votacao_endereco")
		cargo := parseStr(rec, "ds_cargo")
		nrVotavel := parseStr(rec, "nr_votavel")
		nmVotavel := parseStr(rec, "nm_votavel")
		aptos := parseInt(rec, "qt_aptos")
		comp := parseInt(rec, "qt_comparecimento")
		abst := parseInt(rec, "qt_abstencoes")
		nominais := parseInt(rec, "qt_votos_nominais")
		votos := parseInt(rec, "qt_votos")

		_, err = stmt.Exec(zona, local, end, secao, nrLocal, cargo, nrVotavel, nmVotavel, aptos, comp, abst, nominais, votos)
		if err != nil {
			return count, fmt.Errorf("exec insert row %d: %w", count+1, err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return count, nil
}

func (r *Repository) GetSummary() (*model.Summary, error) {
	querySecoes := `
	WITH DistinctSecoes AS (
		SELECT DISTINCT nr_zona, nr_secao, nm_local_votacao, qt_aptos, qt_comparecimento, qt_abstencoes, qt_votos_nominais
		FROM votacao
	)
	SELECT 
		COUNT(*), 
		COUNT(DISTINCT nm_local_votacao),
		COALESCE(SUM(qt_aptos), 0),
		COALESCE(SUM(qt_comparecimento), 0),
		COALESCE(SUM(qt_abstencoes), 0),
		COALESCE(SUM(qt_votos_nominais), 0)
	FROM DistinctSecoes
	`
	var s model.Summary
	err := r.db.QueryRow(querySecoes).Scan(
		&s.TotalSecoes,
		&s.TotalLocais,
		&s.TotalAptos,
		&s.TotalComparecimento,
		&s.TotalAbstencoes,
		&s.TotalVotosNominais,
	)
	if err != nil {
		return nil, fmt.Errorf("query summary secoes: %w", err)
	}

	queryVotos := `
	SELECT 
		COALESCE(SUM(qt_votos), 0),
		COALESCE(SUM(CASE WHEN nr_votavel = '95' OR UPPER(nm_votavel) LIKE '%BRANCO%' THEN qt_votos ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN nr_votavel = '96' OR UPPER(nm_votavel) LIKE '%NULO%' THEN qt_votos ELSE 0 END), 0)
	FROM votacao
	`
	err = r.db.QueryRow(queryVotos).Scan(&s.TotalVotos, &s.TotalBrancos, &s.TotalNulos)
	if err != nil {
		return nil, fmt.Errorf("query summary votos: %w", err)
	}

	return &s, nil
}

func (r *Repository) GetCandidateTotals() ([]model.CandidateTotal, error) {
	summary, err := r.GetSummary()
	if err != nil {
		return nil, err
	}

	totalVotos := summary.TotalVotos
	if totalVotos == 0 {
		totalVotos = 1
	}

	rows, err := r.db.Query(`
		SELECT nr_votavel, nm_votavel, SUM(qt_votos) as total_votos
		FROM votacao
		GROUP BY nr_votavel, nm_votavel
		ORDER BY total_votos DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query candidate totals: %w", err)
	}
	defer rows.Close()

	var list []model.CandidateTotal
	for rows.Next() {
		var c model.CandidateTotal
		if err := rows.Scan(&c.NrVotavel, &c.NmVotavel, &c.Votos); err != nil {
			return nil, err
		}
		c.Percentual = (float64(c.Votos) / float64(totalVotos)) * 100
		list = append(list, c)
	}

	return list, nil
}

func (r *Repository) GetZones() ([]model.ZoneInfo, error) {
	rows, err := r.db.Query(`
		WITH DistinctSecoes AS (
			SELECT DISTINCT nr_zona, nr_secao, nm_local_votacao, qt_aptos
			FROM votacao
		)
		SELECT 
			nr_zona,
			COUNT(DISTINCT nr_secao),
			COUNT(DISTINCT nm_local_votacao),
			COALESCE(SUM(qt_aptos), 0)
		FROM DistinctSecoes
		GROUP BY nr_zona
		ORDER BY nr_zona ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.ZoneInfo
	for rows.Next() {
		var z model.ZoneInfo
		if err := rows.Scan(&z.NrZona, &z.TotalSecoes, &z.TotalLocais, &z.TotalAptos); err != nil {
			return nil, err
		}
		list = append(list, z)
	}
	return list, nil
}

func (r *Repository) SearchVotes(f Filter) ([]model.SectionResult, error) {
	where := []string{"1=1"}
	args := []interface{}{}

	if f.Zona > 0 {
		where = append(where, "nr_zona = ?")
		args = append(args, f.Zona)
	}
	if f.Secao > 0 {
		where = append(where, "nr_secao = ?")
		args = append(args, f.Secao)
	}
	if f.Candidate != "" {
		where = append(where, "(UPPER(nm_votavel) LIKE ? OR nr_votavel = ?)")
		args = append(args, "%"+strings.ToUpper(f.Candidate)+"%", f.Candidate)
	}
	if f.Local != "" {
		where = append(where, "(UPPER(nm_local_votacao) LIKE ? OR UPPER(ds_local_votacao_endereco) LIKE ?)")
		args = append(args, "%"+strings.ToUpper(f.Local)+"%", "%"+strings.ToUpper(f.Local)+"%")
	}

	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT 
			nr_zona, nr_secao, nm_local_votacao, ds_local_votacao_endereco,
			qt_aptos, qt_comparecimento, qt_abstencoes, qt_votos_nominais,
			nr_votavel, nm_votavel, qt_votos
		FROM votacao
		WHERE %s
		ORDER BY qt_votos DESC, nr_zona ASC, nr_secao ASC
		LIMIT ? OFFSET ?
	`, strings.Join(where, " AND "))

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("search votes query: %w", err)
	}
	defer rows.Close()

	var results []model.SectionResult
	for rows.Next() {
		var sr model.SectionResult
		if err := rows.Scan(
			&sr.NrZona, &sr.NrSecao, &sr.NmLocalVotacao, &sr.EnderecoLocal,
			&sr.QtAptos, &sr.QtComparecimento, &sr.QtAbstencoes, &sr.QtVotosNominais,
			&sr.NrVotavel, &sr.NmVotavel, &sr.QtVotos,
		); err != nil {
			return nil, err
		}
		if sr.QtComparecimento > 0 {
			sr.PercentualNaSecao = (float64(sr.QtVotos) / float64(sr.QtComparecimento)) * 100
		}
		results = append(results, sr)
	}

	return results, nil
}

func (r *Repository) GetSectionDetail(zona, secao int) (*model.SectionOverview, error) {
	rows, err := r.db.Query(`
		SELECT 
			nr_zona, nr_secao, nm_local_votacao, ds_local_votacao_endereco,
			qt_aptos, qt_comparecimento, qt_abstencoes, qt_votos_nominais,
			nr_votavel, nm_votavel, qt_votos
		FROM votacao
		WHERE nr_zona = ? AND nr_secao = ?
		ORDER BY qt_votos DESC
	`, zona, secao)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overview *model.SectionOverview
	var totalVotosSecao int

	for rows.Next() {
		var z, s, aptos, comp, abst, nominais, votos int
		var local, end, nrVot, nmVot string

		if err := rows.Scan(&z, &s, &local, &end, &aptos, &comp, &abst, &nominais, &nrVot, &nmVot, &votos); err != nil {
			return nil, err
		}

		if overview == nil {
			overview = &model.SectionOverview{
				NrZona:           z,
				NrSecao:          s,
				NmLocalVotacao:   local,
				EnderecoLocal:    end,
				QtAptos:          aptos,
				QtComparecimento: comp,
				QtAbstencoes:     abst,
				QtVotosNominais:  nominais,
				Resultados:       []model.CandidateResult{},
			}
		}

		totalVotosSecao += votos
		overview.Resultados = append(overview.Resultados, model.CandidateResult{
			NrVotavel: nrVot,
			NmVotavel: nmVot,
			QtVotos:   votos,
		})
	}

	if overview == nil {
		return nil, nil
	}

	overview.TotalVotos = totalVotosSecao
	for i := range overview.Resultados {
		if totalVotosSecao > 0 {
			overview.Resultados[i].Percentual = (float64(overview.Resultados[i].QtVotos) / float64(totalVotosSecao)) * 100
		}
	}

	return overview, nil
}
