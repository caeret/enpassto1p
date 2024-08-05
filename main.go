package main

import (
	"crypto/md5"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/sjson"
	"log"
	"os"
)

var (
	enpassExportedFile      = ""
	onePasswordExportedData = ""
)

const (
	CategoryLogin      = "login"
	CategorySecureNote = "note"
)

func main() {
	b, err := os.ReadFile(enpassExportedFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	newAccounts := make([]any, 0)
	j := jsoniter.Get(b).Get("items")
	for i := 0; i < j.Size(); i++ {
		switch j.Get(i, "category").ToString() {
		case CategoryLogin:
			acc := EnpassAccount{
				UUID:      j.Get(i, "uuid").ToString(),
				Title:     j.Get(i, "title").ToString(),
				SubTitle:  j.Get(i, "subtitle").ToString(),
				CreatedAt: j.Get(i, "createdAt").ToInt64(),
				UpdatedAt: j.Get(i, "updated_at").ToInt64(),
			}
			for k := 0; k < j.Get(i, "fields").Size(); k++ {
				o := j.Get(i, "fields", k)
				switch o.Get("type").ToString() {
				case "username":
					if o.Get("value").ToString() != "" {
						acc.Username = o.Get("value").ToString()
					}
				case "email":
					if o.Get("value").ToString() != "" {
						acc.Email = o.Get("value").ToString()
					}
				case "password":
					if o.Get("value").ToString() != "" {
						acc.Password = o.Get("value").ToString()
					}
				case "url":
					if o.Get("value").ToString() != "" {
						acc.URL = o.Get("value").ToString()
					}
				case "totp":
					if o.Get("value").ToString() != "" {
						acc.TOTP = o.Get("value").ToString()
					}
				}
			}
			if acc.Username == "" {
				acc.Username = acc.Email
			}

			opacc := OnePasswordAccount{
				UUID:         acc.UUID,
				FavIndex:     0,
				CreatedAt:    acc.CreatedAt,
				UpdatedAt:    acc.UpdatedAt,
				State:        "active",
				CategoryUUID: "001",
				Overview: OnePasswordAccountOverview{
					Title:    acc.Title,
					Subtitle: acc.SubTitle,
					Icons:    nil,
					Urls: []OnePasswordAccountOverviewUrl{{
						Label: "网站",
						Url:   acc.URL,
						Mode:  "default",
					}},
					Url:                  acc.URL,
					WatchtowerExclusions: nil,
				},
			}
			opacc.Details.LoginFields = make([]OnePasswordLoginField, 0)
			opacc.Details.PasswordHistory = []interface{}{}
			opacc.Details.Sections = []OnePasswordSection{{
				Title:  "",
				Name:   "add more",
				Fields: make([]OnePasswordSectionField, 0),
			}}
			if acc.Username != "" {
				opacc.Details.LoginFields = append(opacc.Details.LoginFields, OnePasswordLoginField{
					Value:       acc.Username,
					Id:          "",
					Name:        "username",
					FieldType:   "T",
					Designation: "username",
				})
			}
			if acc.Password != "" {
				opacc.Details.LoginFields = append(opacc.Details.LoginFields, OnePasswordLoginField{
					Value:       acc.Password,
					Id:          "",
					Name:        "password",
					FieldType:   "p",
					Designation: "password",
				})
			}

			if acc.Email != "" {
				field := OnePasswordSectionField{
					Title: "电子邮件",
					Id:    "",
					Value: OnePasswordLoginFieldValue{
						Email: &OnePasswordLoginFieldValueEmail{
							EmailAddress: acc.Email,
							Provider:     nil,
						},
					},
					Guarded:      false,
					Multiline:    false,
					DontGenerate: false,
					InputTraits: OnePasswordInputTraits{
						Keyboard:       "emailAddress",
						Correction:     "no",
						Capitalization: "none",
					},
				}
				opacc.Details.Sections[0].Fields = append(opacc.Details.Sections[0].Fields, field)
			}

			if acc.TOTP != "" {
				field := OnePasswordSectionField{
					Title: "一次性密码",
					Id:    "TOTP_" + md5hash(acc.TOTP),
					Value: OnePasswordLoginFieldValue{
						Totp: &acc.TOTP,
					},
					Guarded:      false,
					Multiline:    false,
					DontGenerate: false,
					InputTraits: OnePasswordInputTraits{
						Keyboard:       "default",
						Correction:     "no",
						Capitalization: "none",
					},
				}
				opacc.Details.Sections[0].Fields = append(opacc.Details.Sections[0].Fields, field)
			}

			newAccounts = append(newAccounts, opacc)
		case CategorySecureNote:
			note := EnpassSecureNote{
				UUID:      j.Get(i, "uuid").ToString(),
				Title:     j.Get(i, "title").ToString(),
				SubTitle:  j.Get(i, "subtitle").ToString(),
				Content:   j.Get(i, "note").ToString(),
				CreatedAt: j.Get(i, "createdAt").ToInt64(),
				UpdatedAt: j.Get(i, "updated_at").ToInt64(),
			}

			opacc := OnePasswordAccount{
				UUID:         note.UUID,
				FavIndex:     0,
				CreatedAt:    note.CreatedAt,
				UpdatedAt:    note.UpdatedAt,
				State:        "active",
				CategoryUUID: "003",
				Overview: OnePasswordAccountOverview{
					Title:                note.Title,
					Subtitle:             note.SubTitle,
					Icons:                nil,
					Url:                  "",
					WatchtowerExclusions: nil,
				},
			}
			opacc.Details.LoginFields = make([]OnePasswordLoginField, 0)
			opacc.Details.PasswordHistory = []interface{}{}
			opacc.Details.Sections = make([]OnePasswordSection, 0)
			opacc.Details.NotesPlain = note.Content

			newAccounts = append(newAccounts, opacc)
		}
	}

	b1, err := os.ReadFile(onePasswordExportedData)
	if err != nil {
		panic(err)
	}
	b1, err = sjson.SetBytes(b1, "accounts.0.vaults.0.items", newAccounts)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b1))

	err = os.WriteFile(onePasswordExportedData, b1, 0644)
	if err != nil {
		panic(err)
	}
}

type EnpassAccount struct {
	UUID      string
	Title     string
	SubTitle  string
	Username  string
	Password  string
	Email     string
	TOTP      string
	URL       string
	CreatedAt int64
	UpdatedAt int64
}

type EnpassSecureNote struct {
	UUID      string
	Title     string
	SubTitle  string
	Content   string
	CreatedAt int64
	UpdatedAt int64
}
type OnePasswordSecureNote struct {
	UUID         string `json:"uuid"`
	FavIndex     int64  `json:"favIndex"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
	State        string `json:"state"`
	CategoryUuid string `json:"categoryUuid"`
	Details      struct {
		LoginFields     []interface{} `json:"loginFields"`
		NotesPlain      string        `json:"notesPlain"`
		Sections        []interface{} `json:"sections"`
		PasswordHistory []interface{} `json:"passwordHistory"`
	} `json:"details"`
	Overview struct {
		Subtitle             string      `json:"subtitle"`
		Icons                interface{} `json:"icons"`
		Title                string      `json:"title"`
		Url                  string      `json:"url"`
		WatchtowerExclusions interface{} `json:"watchtowerExclusions"`
	} `json:"overview"`
}

type OnePasswordAccount struct {
	UUID         string                     `json:"uuid"`
	FavIndex     int64                      `json:"favIndex"`
	CreatedAt    int64                      `json:"createdAt"`
	UpdatedAt    int64                      `json:"updatedAt"`
	State        string                     `json:"state"`
	CategoryUUID string                     `json:"categoryUuid"`
	Details      OnePasswordAccountDetails  `json:"details"`
	Overview     OnePasswordAccountOverview `json:"overview"`
}

type OnePasswordAccountOverview struct {
	Title                string                          `json:"title"`
	Subtitle             string                          `json:"subtitle"`
	Icons                interface{}                     `json:"icons"`
	Urls                 []OnePasswordAccountOverviewUrl `json:"urls,omitempty"`
	Url                  string                          `json:"url"`
	WatchtowerExclusions interface{}                     `json:"watchtowerExclusions"`
}

type OnePasswordAccountOverviewUrl struct {
	Label string `json:"label"`
	Url   string `json:"url"`
	Mode  string `json:"mode"`
}

type OnePasswordAccountDetails struct {
	LoginFields     []OnePasswordLoginField `json:"loginFields"`
	NotesPlain      string                  `json:"notesPlain"`
	Sections        []OnePasswordSection    `json:"sections"`
	PasswordHistory []interface{}           `json:"passwordHistory"`
}

type OnePasswordLoginField struct {
	Value       string `json:"value"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	FieldType   string `json:"fieldType"`
	Designation string `json:"designation"`
}

type OnePasswordSection struct {
	Title  string                    `json:"title"`
	Name   string                    `json:"name"`
	Fields []OnePasswordSectionField `json:"fields"`
}

type OnePasswordSectionField struct {
	Title        string                     `json:"title"`
	Id           string                     `json:"id"`
	Value        OnePasswordLoginFieldValue `json:"value"`
	Guarded      bool                       `json:"guarded"`
	Multiline    bool                       `json:"multiline"`
	DontGenerate bool                       `json:"dontGenerate"`
	InputTraits  OnePasswordInputTraits     `json:"inputTraits"`
}

type OnePasswordLoginFieldValue struct {
	Totp  *string                          `json:"totp,omitempty"`
	Email *OnePasswordLoginFieldValueEmail `json:"email,omitempty"`
}

type OnePasswordLoginFieldValueEmail struct {
	EmailAddress string      `json:"email_address"`
	Provider     interface{} `json:"provider"`
}

type OnePasswordInputTraits struct {
	Keyboard       string `json:"keyboard"`
	Correction     string `json:"correction"`
	Capitalization string `json:"capitalization"`
}

func md5hash(in string) string {
	hash := md5.New()
	hash.Write([]byte(in))
	return fmt.Sprintf("%X", hash.Sum(nil))
}
