package repository

import (
	"database/sql"
	"github.com/messwith/coding_challenge/models"
	"github.com/shopspring/decimal"
)

// AccountRepository is essentially set of methods for working with accounts in db
type AccountRepository struct {
	db *sql.DB
}

// NewAccountRepository creates ready to use account repository
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// GetAccounts loads all accounts from db without pagination
func (ar *AccountRepository) GetAccounts() ([]models.Account, error) {

	accounts := []models.Account{}

	rows, err := ar.db.Query(`SELECT id, owner, balance, currency FROM accounts ORDER BY id`)
	if err == sql.ErrNoRows {
		return accounts, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		account := models.Account{}
		err = rows.Scan(&account.ID, &account.Owner, &account.Balance, &account.Currency)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// LockAccount locks specified accounts for updating its balance later
func (ar *AccountRepository) LockAccounts(tx *sql.Tx, senderID, receiverID string) (*models.Account, *models.Account, error) {
	rows, err := tx.Query(`SELECT id, owner, balance, currency FROM accounts 
									WHERE id = $1 OR id = $2 FOR UPDATE`, senderID, receiverID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var sender, receiver *models.Account
	for rows.Next() {
		account := models.Account{}
		if err := rows.Scan(&account.ID, &account.Owner, &account.Balance, &account.Currency); err != nil {
			return nil, nil, err
		}
		if account.ID == senderID {
			sender = &account
		}
		if account.ID == receiverID {
			receiver = &account
		}
	}

	if sender == nil || receiver == nil {
		return nil, nil, sql.ErrNoRows
	}

	return sender, receiver, nil
}

// UpdateAccountBalance updates balance of specified account
func (ar *AccountRepository) UpdateAccountBalance(tx *sql.Tx, accountID string, newBalance decimal.Decimal) (error) {
	_, err := tx.Exec(`UPDATE accounts SET balance = $1 WHERE id = $2`, newBalance, accountID)
	return err
}

