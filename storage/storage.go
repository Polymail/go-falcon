package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/le0pard/go-falcon/log"
	"github.com/le0pard/go-falcon/utils"
	_ "github.com/lib/pq"
	"strconv"
	"strings"
	"time"
)

type StorageConfig struct {
	Adapter   string
	Host      string
	Port      int
	Username  string
	Password  string
	Database  string
	Pool      int
	Pool_Idle int

	Auth_Sql string

	Settings_Sql string

	Messages_Sql    string
	Attachments_Sql string

	Max_Messages_Enabled        bool
	Max_Messages_Cleanup_Sql    string
	Max_Attachments_Cleanup_Sql string

	Spamassassin_Sql string

	Clamav_Sql string

	Pop3_Count_And_Size_Messages string
	Pop3_Messages_List           string
	Pop3_Message_One             string
	Pop3_Message_Delete          string

	Email_Address_Mode_Sql string
}

type DBConn struct {
	DB     *sql.DB
	config *StorageConfig
}

type InboxSettings struct {
	MaxMessages, RateLimit int
}

func InitDatabase(config *StorageConfig) (*DBConn, error) {
	switch strings.ToLower(config.Adapter) {
	case "postgresql":
		db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", config.Username, config.Password, config.Host, config.Port, config.Database))
		if err != nil {
			return nil, err
		}
		db.SetMaxOpenConns(config.Pool)
		db.SetMaxIdleConns(config.Pool_Idle)
		return &DBConn{DB: db, config: config}, err
	default:
		return nil, errors.New("invalid database adapter")
	}
}

// check if user exist

func (db *DBConn) IfUserExist(username string) bool {
	var (
		id       int
		password string
	)
	err := db.DB.QueryRow(db.config.Auth_Sql, username).Scan(&id, &password)
	if err != nil {
		return false
	}
	return true
}

// check username login and return with password

func (db *DBConn) CheckUserWithPass(authMethod, username, cramPassword, cramSecret string) (int, string, error) {
	var (
		id       int
		password string
	)
	log.Debugf("AUTH by %s / %s", username, cramPassword)
	err := db.DB.QueryRow(db.config.Auth_Sql, username).Scan(&id, &password)
	if err != nil {
		log.Debugf("User %s doesn't found (sql should return 'id' and 'password' fields): %v", username, err)
		return 0, "", err
	}
	if !utils.CheckProtocolAuthPass(authMethod, password, cramPassword, cramSecret) {
		log.Debugf("User %s send invalid password", username)
		return 0, "", errors.New("The user have invalid password")
	}
	return id, password, nil
}

// check username login

func (db *DBConn) CheckUser(authMethod, username, cramPassword, cramSecret string) (int, error) {
	id, _, err := db.CheckUserWithPass(authMethod, username, cramPassword, cramSecret)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// check address mode

func (db *DBConn) CheckAddressMode(username string) (int, error) {
	var (
		id int
	)
	log.Debugf("CheckAddressMode by %s", username)
	err := db.DB.QueryRow(db.config.Email_Address_Mode_Sql, username).Scan(&id)
	if err != nil {
		log.Debugf("User Address %s doesn't found in inboxes (sql should return 'id' field): %v", username, err)
		return 0, err
	}
	return id, nil
}

// save email

func (db *DBConn) StoreMail(mailboxId int, subject string, date time.Time, from, from_name, to, to_name, html, text string, rawEmail []byte) (int, error) {
	var (
		id int
	)
	strBody := utils.CheckAndFixUtf8(string(rawEmail))
	sql := strings.Replace(db.config.Messages_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
	// normalize variables
	if len(subject) > 1000 {
		subject = subject[0:1000]
	}
	if len(from) > 255 {
		from = from[0:255]
	}
	if len(from_name) > 255 {
		from_name = from_name[0:255]
	}
	if len(to) > 255 {
		to = to[0:255]
	}
	if len(to_name) > 255 {
		to_name = to_name[0:255]
	}
	// sql
	err := db.DB.QueryRow(sql,
		mailboxId,
		subject,
		date.UTC(),
		from,
		from_name,
		to,
		to_name,
		html,
		text,
		strBody,
		len(strBody)).Scan(&id)
	if err != nil {
		log.Errorf("Messages SQL error: %v", err)
		return 0, err
	}
	if 0 == id {
		log.Errorf("Messages Not return last ID: %v", id)
		return 0, errors.New("Messages Not return last ID")
	}
	return id, nil
}

// update spam report

func (db *DBConn) UpdateSpamReport(mailboxId int, messageId int, spamReport string) (int, error) {
	var (
		id int
	)
	sql := strings.Replace(db.config.Spamassassin_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
	err := db.DB.QueryRow(sql,
		mailboxId,
		messageId,
		spamReport).Scan(&id)
	if err != nil {
		log.Errorf("Spamassassin SQL error: %v", err)
		return 0, err
	}
	return id, nil
}

// update viruses report

func (db *DBConn) UpdateVirusesReport(mailboxId int, messageId int, virusesReport string) (int, error) {
	var (
		id int
	)
	sql := strings.Replace(db.config.Clamav_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
	err := db.DB.QueryRow(sql,
		mailboxId,
		messageId,
		virusesReport).Scan(&id)
	if err != nil {
		log.Errorf("Clamav SQL error: %v", err)
		return 0, err
	}
	return id, nil
}

// save attachment

func (db *DBConn) StoreAttachment(mailboxId int, messageId int, filename, attachmentType, contentType, contentId, transferEncoding, strBody string) (int, error) {
	var (
		id int
	)
	sql := strings.Replace(db.config.Attachments_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
	err := db.DB.QueryRow(sql,
		mailboxId,
		messageId,
		filename,
		attachmentType,
		contentType,
		contentId,
		transferEncoding,
		utils.EncodeBase64(strBody),
		len(strBody)).Scan(&id)
	if err != nil {
		log.Errorf("Attachments SQL error: %v", err)
		return 0, err
	}
	if id == 0 {
		log.Errorf("Attachments Not return last ID: %v", id)
		return 0, errors.New("Attachments Not return last ID")
	}
	return id, nil
}

// get settings

func (db *DBConn) GeInboxSettings(mailboxId int) (InboxSettings, error) {
	var (
		maxMessages int
		rateLimit   int
	)
	err := db.DB.QueryRow(db.config.Settings_Sql, mailboxId).Scan(&maxMessages, &rateLimit)
	if err != nil {
		log.Errorf("Settings SQL error: %v", err)
	}
	return InboxSettings{MaxMessages: maxMessages, RateLimit: rateLimit}, err
}

// cleanup messages
func (db *DBConn) CleanupMessages(mailboxId int, inboxSettings InboxSettings) error {
	if db.config.Max_Messages_Enabled && inboxSettings.MaxMessages > 0 {
		var (
			sql    string
			tmpId  int
			msgIds []string
		)
		sql = strings.Replace(db.config.Max_Messages_Cleanup_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
		rows, err := db.DB.Query(sql, mailboxId, inboxSettings.MaxMessages)
		if err != nil {
			log.Errorf("CleanupMessages SQL error: %v", err)
			return err
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&tmpId)
			if err != nil {
				log.Errorf("CleanupMessages SQL error: %v", err)
				return err
			}
			msgIds = append(msgIds, strconv.Itoa(tmpId))
		}
		if len(msgIds) > 0 {
			sql = strings.Replace(db.config.Max_Attachments_Cleanup_Sql, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
			for _, msgId := range msgIds {
				_, err := db.DB.Exec(sql, mailboxId, msgId)
				if err != nil {
					log.Errorf("CleanupMessages SQL error: %v", err)
					return err
				}
			}
		}
	}
	return nil
}

// pop3 count and sum

func (db *DBConn) Pop3MessagesCountAndSum(mailboxId int) (int, int, error) {
	var (
		sql   string
		count int
		sum   int
	)
	sql = strings.Replace(db.config.Pop3_Count_And_Size_Messages, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
	err := db.DB.QueryRow(sql, mailboxId).Scan(&count, &sum)
	if err != nil {
		log.Debugf("Pop3MessagesCountAndSum SQL error: %v", err) //empty results will be error
		return 0, 0, err
	}
	return count, sum, nil
}

// pop3 messages

func (db *DBConn) Pop3MessagesList(mailboxId int) ([][2]int, error) {
	var (
		sql     string
		tmpId   int
		tmpSize int
		msgIds  [][2]int
	)

	sql = strings.Replace(db.config.Pop3_Messages_List, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
	rows, err := db.DB.Query(sql, mailboxId)
	if err != nil {
		log.Errorf("Pop3MessagesList SQL error: %v", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&tmpId, &tmpSize)
		if err != nil {
			log.Errorf("Pop3MessagesList SQL error: %v", err)
			return nil, err
		}
		msgIds = append(msgIds, [2]int{tmpId, tmpSize})
	}
	return msgIds, nil
}

// pop3 message

func (db *DBConn) Pop3Message(mailboxId, messageId int) (int, string, error) {
	var (
		sql     string
		msgSize int
		msgBody string
	)

	sql = strings.Replace(db.config.Pop3_Message_One, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
	err := db.DB.QueryRow(sql, mailboxId, messageId).Scan(&msgSize, &msgBody)
	if err != nil {
		log.Debugf("Pop3Message SQL error: %v", err)
		return 0, "", err
	}
	return msgSize, msgBody, nil
}

// pop3 delete message

func (db *DBConn) Pop3DeleteMessage(mailboxId, messageId int) error {
	var (
		sql   string
		retId int
	)

	sql = strings.Replace(db.config.Pop3_Message_Delete, "[[inbox_id]]", strconv.Itoa(mailboxId), 1)
	err := db.DB.QueryRow(sql, mailboxId, messageId).Scan(&retId)
	if err != nil {
		log.Debugf("Pop3DeleteMessage SQL error: %v", err)
		return err
	}
	return nil
}

// close connection

func (db *DBConn) Close() {
	db.DB.Close()
}
