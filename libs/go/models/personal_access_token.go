package models

import "time"

// PersonalAccessToken stores the active AI Kitchen Pass state for one user.
type PersonalAccessToken struct {
	UserID         string     `firestore:"userID" json:"userID"`
	TokenDigest    string     `firestore:"tokenDigest" json:"tokenDigest"`
	EncryptedToken string     `firestore:"encryptedToken" json:"encryptedToken"`
	TokenPreview   string     `firestore:"tokenPreview" json:"tokenPreview"`
	CreatedDate    time.Time  `firestore:"createdDate" json:"createdDate"`
	RotatedDate    *time.Time `firestore:"rotatedDate,omitempty" json:"rotatedDate,omitempty"`
	RevokedDate    *time.Time `firestore:"revokedDate,omitempty" json:"revokedDate,omitempty"`
	LastUsedDate   *time.Time `firestore:"lastUsedDate,omitempty" json:"lastUsedDate,omitempty"`
	LastUsedAction *string    `firestore:"lastUsedAction,omitempty" json:"lastUsedAction,omitempty"`
}
