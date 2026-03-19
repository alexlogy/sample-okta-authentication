package models

type Config struct {
	AppURL             string `json:"app_url"`
	SAMLIDPMetadataURL string `json:"saml_idp_metadata_url"`
	SAMLSPCertFile     string `json:"saml_sp_cert_file"`
	SAMLSPKeyFile      string `json:"saml_sp_key_file"`
	LogoutURL          string `json:"logout_url"`
}
