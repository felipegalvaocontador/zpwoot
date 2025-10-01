package uuid

import (
	"github.com/google/uuid"
)

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Generate() string {
	return uuid.New().String()
}

func (g *Generator) GenerateShort() string {
	return uuid.New().String()[:8]
}

func (g *Generator) Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func (g *Generator) IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

var defaultGenerator = New()

func Generate() string {
	return defaultGenerator.Generate()
}

func GenerateShort() string {
	return defaultGenerator.GenerateShort()
}

func Parse(s string) (uuid.UUID, error) {
	return defaultGenerator.Parse(s)
}

func IsValid(s string) bool {
	return defaultGenerator.IsValid(s)
}
