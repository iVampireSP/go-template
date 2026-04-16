package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Mixin of the User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			MaxLen(255).
			NotEmpty().
			Unique(),
		field.String("password_hash").
			MaxLen(255).
			Optional().
			Nillable().
			Sensitive(),
		field.String("display_name").
			MaxLen(200).
			NotEmpty(),
		field.String("avatar_url").
			MaxLen(500).
			Optional().
			Nillable(),
		field.Bool("email_verified").
			Default(false),
		field.Time("email_verified_at").
			Optional().
			Nillable(),
		field.Enum("status").
			Values("active", "inactive", "suspended", "banned").
			Default("active"),
		field.Time("last_login_at").
			Optional().
			Nillable(),
		field.String("last_login_ip").
			MaxLen(45).
			Optional().
			Nillable(),
		field.String("register_ip").
			MaxLen(45).
			Optional().
			Nillable(),
		field.Int("login_count").
			Default(0),
		field.Time("failed_login_at").
			Optional().
			Nillable(),
		field.Int("failed_login_count").
			Default(0),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}

// Annotations of the User.
func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "users"},
	}
}
