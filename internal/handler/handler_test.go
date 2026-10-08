package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"consulta-resultado-votacao-presidente-2026-oz-gemini/internal/data"
	"consulta-resultado-votacao-presidente-2026-oz-gemini/internal/handler"
	"consulta-resultado-votacao-presidente-2026-oz-gemini/internal/model"
)

const testCSV = `aa_eleicao;cd_tipo_eleicao;nm_tipo_eleicao;cd_eleicao;ds_eleicao;dt_eleicao;sg_uf;cd_municipio;nm_municipio;nr_zona;nm_local_votacao;ds_local_votacao_endereco;nr_secao;nr_local_votacao;cd_modelo_urna;ds_modelo_urna;nr_turno;ds_cargo;nr_votavel;nm_votavel;sq_candidato;qt_aptos;qt_comparecimento;qt_abstencoes;qt_votos_nominais;qt_votos;dt_carga;qt_registros
2026;2;Eleicao Ordinaria;6257;Eleicao Geral Federal 2026;2026-10-04 00:00:00;SP;67890;OSASCO;315;EMEF OSVALDO;RUA ANTONIO, 394;74;1066;2020;UE 2020;1;Presidente;13;LUIZ INACIO LULA DA SILVA;280002542548;256;174;82;161;100;2026-10-06 14:03:43;1
2026;2;Eleicao Ordinaria;6257;Eleicao Geral Federal 2026;2026-10-04 00:00:00;SP;67890;OSASCO;315;EMEF OSVALDO;RUA ANTONIO, 394;74;1066;2020;UE 2020;1;Presidente;22;FLAVIO NANTES BOLSONARO;280002551544;256;174;82;161;61;2026-10-06 14:03:43;1
`

func setupServer(t *testing.T) http.Handler {
	repo, err := data.NewRepository(":memory:")
	if err != nil {
		t.Fatalf("failed repo: %v", err)
	}
	_, err = repo.ImportCSV(strings.NewReader(testCSV))
	if err != nil {
		t.Fatalf("failed import: %v", err)
	}

	h := handler.NewHandler(repo)
	return h.Routes()
}

func TestAPISummary(t *testing.T) {
	mux := setupServer(t)

	req := httptest.NewRequest("GET", "/api/summary", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var sum model.Summary
	if err := json.Unmarshal(rec.Body.Bytes(), &sum); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if sum.TotalSecoes != 1 {
		t.Errorf("expected 1 secao, got %d", sum.TotalSecoes)
	}
	if sum.TotalVotos != 161 {
		t.Errorf("expected 161 votos, got %d", sum.TotalVotos)
	}
}

func TestAPICandidates(t *testing.T) {
	mux := setupServer(t)

	req := httptest.NewRequest("GET", "/api/candidates", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var list []model.CandidateTotal
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(list))
	}
}

func TestAPISearch(t *testing.T) {
	mux := setupServer(t)

	req := httptest.NewRequest("GET", "/api/search?candidate=LULA&zona=315&secao=74", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var results []model.SectionResult
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].NmVotavel != "LUIZ INACIO LULA DA SILVA" {
		t.Errorf("unexpected candidate: %s", results[0].NmVotavel)
	}
}

func TestAPISectionDetail(t *testing.T) {
	mux := setupServer(t)

	req := httptest.NewRequest("GET", "/api/section?zona=315&secao=74", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var overview model.SectionOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if overview.NrSecao != 74 {
		t.Errorf("expected secao 74, got %d", overview.NrSecao)
	}
	if len(overview.Resultados) != 2 {
		t.Errorf("expected 2 results, got %d", len(overview.Resultados))
	}
}
