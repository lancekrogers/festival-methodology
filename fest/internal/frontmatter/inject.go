package frontmatter

import (
	"bytes"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Inject adds frontmatter to content
func Inject(content []byte, fm *Frontmatter) ([]byte, error) {
	fmBytes, err := Marshal(fm)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n\n")
	buf.Write(content)

	return buf.Bytes(), nil
}

// Marshal converts frontmatter to YAML bytes
func Marshal(fm *Frontmatter) ([]byte, error) {
	return yaml.Marshal(fm)
}

// Format returns formatted frontmatter as a string
func Format(fm *Frontmatter) (string, error) {
	data, err := Marshal(fm)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatBlock returns frontmatter wrapped in delimiters
func FormatBlock(fm *Frontmatter) (string, error) {
	data, err := Format(fm)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("---\n%s---\n", data), nil
}

// InjectString adds frontmatter to content string
func InjectString(content string, fm *Frontmatter) (string, error) {
	result, err := Inject([]byte(content), fm)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// MergeInto merges new frontmatter values into existing frontmatter
func MergeInto(existing, new *Frontmatter) *Frontmatter {
	result := *existing

	// Update only non-zero values from new
	if new.Type != "" {
		result.Type = new.Type
	}
	if new.ID != "" {
		result.ID = new.ID
	}
	if new.Name != "" {
		result.Name = new.Name
	}
	if new.Parent != "" {
		result.Parent = new.Parent
	}
	if new.Order != 0 {
		result.Order = new.Order
	}
	if new.Status != "" {
		result.Status = new.Status
	}
	if new.Priority != "" {
		result.Priority = new.Priority
	}
	if new.Autonomy != "" {
		result.Autonomy = new.Autonomy
	}
	if new.GateType != "" {
		result.GateType = new.GateType
	}
	if new.Managed {
		result.Managed = new.Managed
	}
	if len(new.Tags) > 0 {
		result.Tags = new.Tags
	}
	if !new.Created.IsZero() {
		result.Created = new.Created
	}
	if !new.Updated.IsZero() {
		result.Updated = new.Updated
	}

	return &result
}

// Builder provides a fluent interface for constructing frontmatter
type Builder struct {
	fm *Frontmatter
}

// NewBuilder creates a new frontmatter builder
func NewBuilder(docType Type) *Builder {
	return &Builder{
		fm: &Frontmatter{
			Type:   docType,
			Status: DefaultStatus(docType),
		},
	}
}

// ID sets the document ID
func (b *Builder) ID(id string) *Builder {
	b.fm.ID = id
	return b
}

// Name sets the document name
func (b *Builder) Name(name string) *Builder {
	b.fm.Name = name
	return b
}

// Parent sets the parent reference
func (b *Builder) Parent(parent string) *Builder {
	b.fm.Parent = parent
	return b
}

// Order sets the numeric order
func (b *Builder) Order(order int) *Builder {
	b.fm.Order = order
	return b
}

// Status sets the status
func (b *Builder) Status(status Status) *Builder {
	b.fm.Status = status
	return b
}

// Priority sets the priority (festivals only)
func (b *Builder) Priority(priority Priority) *Builder {
	b.fm.Priority = priority
	return b
}

// Autonomy sets the autonomy level (tasks only)
func (b *Builder) Autonomy(autonomy Autonomy) *Builder {
	b.fm.Autonomy = autonomy
	return b
}

// GateType sets the gate type (gates only)
func (b *Builder) GateType(gateType GateType) *Builder {
	b.fm.GateType = gateType
	return b
}

// Managed sets the managed flag
func (b *Builder) Managed(managed bool) *Builder {
	b.fm.Managed = managed
	return b
}

// Tags sets the tags
func (b *Builder) Tags(tags ...string) *Builder {
	b.fm.Tags = tags
	return b
}

// Created sets the creation timestamp
func (b *Builder) Created(t time.Time) *Builder {
	b.fm.Created = t
	return b
}

// Now sets the creation timestamp to now
func (b *Builder) Now() *Builder {
	b.fm.Created = time.Now()
	return b
}

// Build returns the constructed frontmatter
func (b *Builder) Build() *Frontmatter {
	return b.fm
}
