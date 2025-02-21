package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// BaseMixin 实现了基础字段和 ID 类型
type BaseMixin struct {
	mixin.Schema
}

// Fields of the BaseMixin.
func (BaseMixin) Fields() []ent.Field {
	return []ent.Field{
		//.GoType(dto.EntityId(0))
		field.Uint64("id").Positive().
			Immutable().
			Comment("实体 ID"),
	}
}

//// IDType 自定义 ID 类型为 dto.EntityId
//func (BaseMixin) IDType() map[string]func(...field.Descriptor) ent.Field {
//	return map[string]func(...field.Descriptor) ent.Field{
//		"id": func(descriptors ...field.Descriptor) ent.Field {
//			return field.Uint64("id").
//				GoType(dto.EntityId(0)).
//				Positive().
//				Immutable().
//				Comment("实体 ID")
//		},
//	}
//}
