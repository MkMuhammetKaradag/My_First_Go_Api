package dto

// ActivationRequest kullanıcı aktivasyonu için gerekli verileri tutacak struct
type ActivationRequest struct {
	ActivationToken string `json:"activationToken"`
	ActivationCode  string `json:"activationCode"`
}
