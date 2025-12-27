package festival

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// RemoveElement removes an element and renumbers subsequent ones
func (r *Renumberer) RemoveElement(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("Renumberer.RemoveElement")
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)

	// Determine element type
	var elemType ElementType
	var elements []FestivalElement
	var err error

	if matched := regexp.MustCompile(`^\d{3}_`).MatchString(base); matched {
		elemType = PhaseType
		elements, err = r.parser.ParsePhases(ctx, dir)
	} else if matched := regexp.MustCompile(`^\d{2}_`).MatchString(base); matched {
		if strings.HasSuffix(base, ".md") {
			elemType = TaskType
			elements, err = r.parser.ParseTasks(ctx, dir)
		} else {
			elemType = SequenceType
			elements, err = r.parser.ParseSequences(ctx, dir)
		}
	} else {
		return errors.Validation("unable to determine element type").
			WithOp("Renumberer.RemoveElement").
			WithField("path", path)
	}

	if err != nil {
		return errors.Wrap(err, "failed to parse elements").
			WithOp("Renumberer.RemoveElement").
			WithCode(errors.ErrCodeParse)
	}

	// Find element to remove
	var toRemove *FestivalElement
	var removeIndex int
	for i, elem := range elements {
		if elem.Path == path {
			toRemove = &elem
			removeIndex = i
			break
		}
	}

	if toRemove == nil {
		return errors.NotFound("element").WithField("path", path)
	}

	// Build changes
	r.changes = []Change{
		{
			Type:    ChangeRemove,
			OldPath: path,
		},
	}

	// Renumber subsequent elements
	for i := removeIndex + 1; i < len(elements); i++ {
		newNumber := elements[i].Number - 1
		newName := BuildElementName(newNumber, elements[i].Name, elemType)
		if elemType == TaskType {
			newName += ".md"
		}
		newPath := filepath.Join(dir, newName)

		r.changes = append(r.changes, Change{
			Type:    ChangeRename,
			OldPath: elements[i].Path,
			NewPath: newPath,
			Element: elements[i],
		})
	}

	return r.executeChanges()
}
