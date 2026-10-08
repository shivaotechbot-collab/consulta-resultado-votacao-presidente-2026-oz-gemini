package model

type Summary struct {
	TotalSecoes       int `json:"total_secoes"`
	TotalLocais       int `json:"total_locais"`
	TotalAptos        int `json:"total_aptos"`
	TotalComparecimento int `json:"total_comparecimento"`
	TotalAbstencoes   int `json:"total_abstencoes"`
	TotalVotos        int `json:"total_votos"`
	TotalVotosNominais int `json:"total_votos_nominais"`
	TotalBrancos      int `json:"total_brancos"`
	TotalNulos        int `json:"total_nulos"`
}

type CandidateTotal struct {
	NrVotavel  string  `json:"nr_votavel"`
	NmVotavel  string  `json:"nm_votavel"`
	Votos      int     `json:"votos"`
	Percentual float64 `json:"percentual"`
}

type SectionResult struct {
	NrZona            int    `json:"nr_zona"`
	NrSecao           int    `json:"nr_secao"`
	NmLocalVotacao    string `json:"nm_local_votacao"`
	EnderecoLocal     string `json:"ds_local_votacao_endereco"`
	QtAptos           int    `json:"qt_aptos"`
	QtComparecimento  int    `json:"qt_comparecimento"`
	QtAbstencoes      int    `json:"qt_abstencoes"`
	QtVotosNominais   int    `json:"qt_votos_nominais"`
	NrVotavel         string `json:"nr_votavel"`
	NmVotavel         string `json:"nm_votavel"`
	QtVotos           int    `json:"qt_votos"`
	PercentualNaSecao float64 `json:"percentual_na_secao"`
}

type SectionOverview struct {
	NrZona           int               `json:"nr_zona"`
	NrSecao          int               `json:"nr_secao"`
	NmLocalVotacao   string            `json:"nm_local_votacao"`
	EnderecoLocal    string            `json:"ds_local_votacao_endereco"`
	QtAptos          int               `json:"qt_aptos"`
	QtComparecimento int               `json:"qt_comparecimento"`
	QtAbstencoes     int               `json:"qt_abstencoes"`
	QtVotosNominais  int               `json:"qt_votos_nominais"`
	TotalVotos       int               `json:"total_votos"`
	Resultados       []CandidateResult `json:"resultados"`
}

type CandidateResult struct {
	NrVotavel  string  `json:"nr_votavel"`
	NmVotavel  string  `json:"nm_votavel"`
	QtVotos    int     `json:"qt_votos"`
	Percentual float64 `json:"percentual"`
}

type ZoneInfo struct {
	NrZona      int `json:"nr_zona"`
	TotalSecoes int `json:"total_secoes"`
	TotalLocais int `json:"total_locais"`
	TotalAptos  int `json:"total_aptos"`
}
