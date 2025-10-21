package models

// GetAllModels returns all model types for migration
func GetAllModels() []interface{} {
	return []interface{}{
		// User models
		&User{},
		&UserPreference{},
		&Session{},
		&APIToken{},
		&UserFollow{},
		&UserBlock{},
		
		// Gist models
		&Gist{},
		&GistFile{},
		&GistStar{},
		&GistComment{},
		&GistView{},
		&GistWatch{},
		
		// Organization models
		&Organization{},
		&OrganizationMember{},
		&OrganizationInvitation{},
		
		// Transfer models
		&TransferRequest{},
		&TransferHistory{},
		
		// Compliance models
		&AuditLog{},
		&SecurityEvent{},
		&ComplianceLog{},
		&DataRetentionPolicy{},
		&GDPRExportRequest{},
		&GDPRDeletionRequest{},
		
		// Webhook models
		&Webhook{},
		&WebhookDelivery{},
		&WebhookSubscription{},
		
		// Backup models
		&Backup{},
		
		// Migration models
		&Migration{},
		&ImportJob{},
		&ImportItem{},
		
		// System models
		&SystemConfig{},
		
		// Search models
		&SavedSearch{},
		&SearchHistory{},
		&SearchIndexMetadata{},

		// Social models
		&ActivityFeed{},
		&ActivityFeedFollow{},
		&ActivityFeedSubscription{},
		
		// Tag models
		&Tag{},
		&GistTag{},
		
		// Custom domain models
		&CustomDomain{},
	}
}