package models

// GetAllModels returns all model types for migration
func GetAllModels() []interface{} {
	return []interface{}{
		&User{},
		&Session{},
		&APIToken{},
		&WebAuthnCredential{},
		&Follow{},
		&Gist{},
		&GistFile{},
		&Star{},
		&Comment{},
		&Organization{},
		&OrganizationUser{},
		&Team{},
		&TeamMember{},
		&Transfer{},
		&AuditLog{},
		&Webhook{},
		&WebhookDelivery{},
		&Backup{},
		&ImportJob{},
		&SystemConfig{},
	}
}