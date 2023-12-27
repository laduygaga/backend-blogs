// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// ContactsColumns holds the columns for the "contacts" table.
	ContactsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "email", Type: field.TypeString},
		{Name: "link", Type: field.TypeString},
		{Name: "type", Type: field.TypeString},
		{Name: "message", Type: field.TypeString},
		{Name: "created_at", Type: field.TypeTime},
	}
	// ContactsTable holds the schema information for the "contacts" table.
	ContactsTable = &schema.Table{
		Name:       "contacts",
		Columns:    ContactsColumns,
		PrimaryKey: []*schema.Column{ContactsColumns[0]},
	}
	// PasswordTokensColumns holds the columns for the "password_tokens" table.
	PasswordTokensColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "hash", Type: field.TypeString},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "password_token_user", Type: field.TypeInt},
	}
	// PasswordTokensTable holds the schema information for the "password_tokens" table.
	PasswordTokensTable = &schema.Table{
		Name:       "password_tokens",
		Columns:    PasswordTokensColumns,
		PrimaryKey: []*schema.Column{PasswordTokensColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "password_tokens_users_user",
				Columns:    []*schema.Column{PasswordTokensColumns[3]},
				RefColumns: []*schema.Column{UsersColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// PostsColumns holds the columns for the "posts" table.
	PostsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "title", Type: field.TypeString},
		{Name: "body", Type: field.TypeString},
		{Name: "author", Type: field.TypeString},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
	}
	// PostsTable holds the schema information for the "posts" table.
	PostsTable = &schema.Table{
		Name:       "posts",
		Columns:    PostsColumns,
		PrimaryKey: []*schema.Column{PostsColumns[0]},
	}
	// UsersColumns holds the columns for the "users" table.
	UsersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "name", Type: field.TypeString},
		{Name: "email", Type: field.TypeString, Unique: true},
		{Name: "permission", Type: field.TypeString, Default: "Viewer"},
		{Name: "password", Type: field.TypeString},
		{Name: "verified", Type: field.TypeBool, Default: true},
		{Name: "created_at", Type: field.TypeTime},
	}
	// UsersTable holds the schema information for the "users" table.
	UsersTable = &schema.Table{
		Name:       "users",
		Columns:    UsersColumns,
		PrimaryKey: []*schema.Column{UsersColumns[0]},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		ContactsTable,
		PasswordTokensTable,
		PostsTable,
		UsersTable,
	}
)

func init() {
	PasswordTokensTable.ForeignKeys[0].RefTable = UsersTable
}
