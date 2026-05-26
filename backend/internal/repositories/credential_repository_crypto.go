package repositories

import (
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils"
)

// encryptSecrets encrypts plaintext secret fields on a CredentialConnection.
func (r *CredentialRepository) encryptSecrets(conn *models.CredentialConnection) error {
	if r.encryption == nil {
		return nil
	}
	var err error
	if conn.AccessToken != "" && !utils.IsEncrypted(conn.AccessToken) {
		conn.AccessToken, err = r.encryption.Encrypt(conn.AccessToken)
		if err != nil {
			return fmt.Errorf("encrypt access_token: %w", err)
		}
	}
	if conn.RefreshToken != "" && !utils.IsEncrypted(conn.RefreshToken) {
		conn.RefreshToken, err = r.encryption.Encrypt(conn.RefreshToken)
		if err != nil {
			return fmt.Errorf("encrypt refresh_token: %w", err)
		}
	}
	if conn.ShopCipher != "" && !utils.IsEncrypted(conn.ShopCipher) {
		conn.ShopCipher, err = r.encryption.Encrypt(conn.ShopCipher)
		if err != nil {
			return fmt.Errorf("encrypt shop_cipher: %w", err)
		}
	}
	return nil
}

// decryptSecrets decrypts Fernet-encrypted secret fields on a CredentialConnection.
func (r *CredentialRepository) decryptSecrets(conn *models.CredentialConnection) error {
	if r.encryption == nil {
		return nil
	}
	var err error
	if conn.AccessToken != "" && utils.IsEncrypted(conn.AccessToken) {
		conn.AccessToken, err = r.encryption.Decrypt(conn.AccessToken)
		if err != nil {
			return fmt.Errorf("decrypt access_token: %w", err)
		}
	}
	if conn.RefreshToken != "" && utils.IsEncrypted(conn.RefreshToken) {
		conn.RefreshToken, err = r.encryption.Decrypt(conn.RefreshToken)
		if err != nil {
			return fmt.Errorf("decrypt refresh_token: %w", err)
		}
	}
	if conn.ShopCipher != "" && utils.IsEncrypted(conn.ShopCipher) {
		conn.ShopCipher, err = r.encryption.Decrypt(conn.ShopCipher)
		if err != nil {
			return fmt.Errorf("decrypt shop_cipher: %w", err)
		}
	}
	return nil
}

// encryptAppConfigSecrets encrypts plaintext secret fields on a CredentialAppConfig.
func (r *CredentialRepository) encryptAppConfigSecrets(cfg *models.CredentialAppConfig) error {
	if r.encryption == nil {
		return nil
	}
	var err error
	if cfg.AppKey != "" && !utils.IsEncrypted(cfg.AppKey) {
		cfg.AppKey, err = r.encryption.Encrypt(cfg.AppKey)
		if err != nil {
			return fmt.Errorf("encrypt app_key: %w", err)
		}
	}
	if cfg.AppSecret != "" && !utils.IsEncrypted(cfg.AppSecret) {
		cfg.AppSecret, err = r.encryption.Encrypt(cfg.AppSecret)
		if err != nil {
			return fmt.Errorf("encrypt app_secret: %w", err)
		}
	}
	if cfg.PartnerKey != "" && !utils.IsEncrypted(cfg.PartnerKey) {
		cfg.PartnerKey, err = r.encryption.Encrypt(cfg.PartnerKey)
		if err != nil {
			return fmt.Errorf("encrypt partner_key: %w", err)
		}
	}
	return nil
}

// decryptAppConfigSecrets decrypts Fernet-encrypted secret fields on a CredentialAppConfig.
func (r *CredentialRepository) decryptAppConfigSecrets(cfg *models.CredentialAppConfig) error {
	if r.encryption == nil {
		return nil
	}
	var err error
	if cfg.AppKey != "" && utils.IsEncrypted(cfg.AppKey) {
		cfg.AppKey, err = r.encryption.Decrypt(cfg.AppKey)
		if err != nil {
			return fmt.Errorf("decrypt app_key: %w", err)
		}
	}
	if cfg.AppSecret != "" && utils.IsEncrypted(cfg.AppSecret) {
		cfg.AppSecret, err = r.encryption.Decrypt(cfg.AppSecret)
		if err != nil {
			return fmt.Errorf("decrypt app_secret: %w", err)
		}
	}
	if cfg.PartnerKey != "" && utils.IsEncrypted(cfg.PartnerKey) {
		cfg.PartnerKey, err = r.encryption.Decrypt(cfg.PartnerKey)
		if err != nil {
			return fmt.Errorf("decrypt partner_key: %w", err)
		}
	}
	return nil
}
