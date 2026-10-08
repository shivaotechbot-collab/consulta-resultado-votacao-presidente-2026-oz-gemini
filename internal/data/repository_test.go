package data_test

import (
	"strings"
	"testing"

	"consulta-resultado-votacao-presidente-2026-oz-gemini/internal/data"
)

const sampleCSV = `aa_eleicao;cd_tipo_eleicao;nm_tipo_eleicao;cd_eleicao;ds_eleicao;dt_eleicao;sg_uf;cd_municipio;nm_municipio;nr_zona;nm_local_votacao;ds_local_votacao_endereco;nr_secao;nr_local_votacao;cd_modelo_urna;ds_modelo_urna;nr_turno;ds_cargo;nr_votavel;nm_votavel;sq_candidato;qt_aptos;qt_comparecimento;qt_abstencoes;qt_votos_nominais;qt_votos;dt_carga;qt_registros
2026;2;Eleicao Ordinaria;6257;Eleicao Geral Federal 2026;2026-10-04 00:00:00;SP;67890;OSASCO;315;EMEF OSVALDO QUIRINO;RUA ANTONIO BENEDITO, 394;74;1066;2020;UE 2020;1;Presidente;13;LUIZ INACIO LULA DA SILVA;280002542548;256;174;82;161;100;2026-10-06 14:03:43;1
2026;2;Eleicao Ordinaria;6257;Eleicao Geral Federal 2026;2026-10-04 00:00:00;SP;67890;OSASCO;315;EMEF OSVALDO QUIRINO;RUA ANTONIO BENEDITO, 394;74;1066;2020;UE 2020;1;Presidente;22;FLAVIO NANTES BOLSONARO;280002551544;256;174;82;161;61;2026-10-06 14:03:43;1
2026;2;Eleicao Ordinaria;6257;Eleicao Geral Federal 2026;2026-10-04 00:00:00;SP;67890;OSASCO;315;EMEF OSVALDO QUIRINO;RUA ANTONIO BENEDITO, 394;74;1066;2020;UE 2020;1;Presidente;95;VOTO BRANCO;-1;256;174;82;161;5;2026-10-06 14:03:43;1
2026;2;Eleicao Ordinaria;6257;Eleicao Geral Federal 2026;2026-10-04 00:00:00;SP;67890;OSASCO;315;EMEF OSVALDO QUIRINO;RUA ANTONIO BENEDITO, 394;74;1066;2020;UE 2020;1;Presidente;96;VOTO NULO;-1;256;174;82;161;8;2026-10-06 14:03:43;1
2026;2;Eleicao Ordinaria;6257;Eleicao Geral Federal 2026;2026-10-04 00:00:00;SP;67890;OSASCO;213;EE PROF VICENTE;RUA DIOGO, 03;84;1252;2022;UE 2022;1;Presidente;13;LUIZ INACIO LULA DA SILVA;280002542548;300;200;100;190;120;2026-10-06 14:03:43;1
2026;2;Eleicao Ordinaria;6257;Eleicao Geral Federal 2026;2026-10-04 00:00:00;SP;67890;OSASCO;213;EE PROF VICENTE;RUA DIOGO, 03;84;1252;2022;UE 2022;1;Presidente;22;FLAVIO NANTES BOLSONARO;280002551544;300;200;100;190;70;2026-10-06 14:03:43;1
2026;2;Eleicao Ordinaria;6257;Eleicao Geral Federal 2026;2026-10-04 00:00:00;SP;67890;OSASCO;213;EE PROF VICENTE;RUA DIOGO, 03;84;1252;2022;UE 2022;1;Presidente;95;VOTO BRANCO;-1;300;200;100;190;4;2026-10-06 14:03:43;1
2026;2;Eleicao Ordinaria;6257;Eleicao Geral Federal 2026;2026-10-04 00:00:00;SP;67890;OSASCO;213;EE PROF VICENTE;RUA DIOGO, 03;84;1252;2022;UE 2022;1;Presidente;96;VOTO NULO;-1;300;200;100;190;6;2026-10-06 14:03:43;1
`

func TestInitAndQueryDB(t *testing.T) {
	repo, err := data.NewRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	defer repo.Close()

	r := strings.NewReader(sampleCSV)
	n, err := repo.ImportCSV(r)
	if err != nil {
		t.Fatalf("failed to import csv: %v", err)
	}
	if n != 8 {
		t.Fatalf("expected 8 rows imported, got %d", n)
	}

	summary, err := repo.GetSummary()
	if err != nil {
		t.Fatalf("failed to get summary: %v", err)
	}
	if summary.TotalSecoes != 2 {
		t.Errorf("expected 2 secoes, got %d", summary.TotalSecoes)
	}
	if summary.TotalAptos != 556 {
		t.Errorf("expected 556 aptos, got %d", summary.TotalAptos)
	}

	candidates, err := repo.GetCandidateTotals()
	if err != nil {
		t.Fatalf("failed to get candidates: %v", err)
	}
	if len(candidates) != 4 {
		t.Fatalf("expected 4 candidates/options, got %d", len(candidates))
	}
	if candidates[0].NmVotavel != "LUIZ INACIO LULA DA SILVA" || candidates[0].Votos != 220 {
		t.Errorf("unexpected top candidate: %+v", candidates[0])
	}

	// Search by candidate name
	results, err := repo.SearchVotes(data.Filter{Candidate: "LULA"})
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 section results for LULA, got %d", len(results))
	}

	// Search by zona and secao
	secOverview, err := repo.GetSectionDetail(315, 74)
	if err != nil {
		t.Fatalf("failed to get section detail: %v", err)
	}
	if secOverview == nil {
		t.Fatal("expected section overview not nil")
	}
	if secOverview.QtAptos != 256 {
		t.Errorf("expected 256 aptos, got %d", secOverview.QtAptos)
	}
	if len(secOverview.Resultados) != 4 {
		t.Errorf("expected 4 candidate rows in section 74, got %d", len(secOverview.Resultados))
	}
}
