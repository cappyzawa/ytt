package template

import (
	"fmt"
	"strings"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/files"
	"github.com/spf13/cobra"
)

type RegularFilesSourceOpts struct {
	files               []string
	fileMarks           []string
	filterTemplateFiles []string
	recursive           bool
	output              string
}

func (s *RegularFilesSourceOpts) Set(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&s.files, "file", "f", nil, "File (ie local path, HTTP URL, -) (can be specified multiple times)")
	cmd.Flags().StringSliceVar(&s.fileMarks, "file-mark", nil, "File mark (ie change file path, mark as non-template) (format: file:key=value) (can be specified multiple times)")
	cmd.Flags().StringSliceVar(&s.filterTemplateFiles, "filter-template-file", nil, "Specify which file to template (can be specified multiple times)")
	cmd.Flags().BoolVarP(&s.recursive, "recursive", "R", false, "Interpret file as directory")
	cmd.Flags().StringVarP(&s.output, "output", "o", "", "Directory for output")
}

type RegularFilesSource struct {
	opts RegularFilesSourceOpts
	ui   cmdcore.PlainUI
}

func NewRegularFilesSource(opts RegularFilesSourceOpts, ui cmdcore.PlainUI) *RegularFilesSource {
	return &RegularFilesSource{opts, ui}
}

func (s *RegularFilesSource) HasInput() bool  { return len(s.opts.files) > 0 }
func (s *RegularFilesSource) HasOutput() bool { return true }

func (s *RegularFilesSource) Input() (TemplateInput, error) {
	filesToProcess, err := files.NewFiles(s.opts.files, s.opts.recursive)
	if err != nil {
		return TemplateInput{}, err
	}

	// Mark some files as non template files
	if len(s.opts.filterTemplateFiles) > 0 {
		for _, file := range filesToProcess {
			var isTemplate bool
			for _, filteredFile := range s.opts.filterTemplateFiles {
				if filteredFile == file.RelativePath() {
					isTemplate = true
					break
				}
			}
			if !isTemplate {
				file.MarkTemplate(false)
			}
		}
	}

	err = s.applyFileMarks(filesToProcess)
	if err != nil {
		return TemplateInput{}, err
	}

	return TemplateInput{Files: filesToProcess}, nil
}

func (s *RegularFilesSource) Output(out TemplateOutput) error {
	if out.Err != nil {
		return out.Err
	}

	if len(s.opts.output) > 0 {
		return files.NewOutputDirectory(s.opts.output, out.Files, s.ui).Write()
	}

	combinedDocBytes, err := out.DocSet.AsBytes()
	if err != nil {
		return fmt.Errorf("Marshaling combined template result: %s", err)
	}

	s.ui.Debugf("### result\n")
	s.ui.Printf("%s", combinedDocBytes) // no newline

	return nil
}

func (s *RegularFilesSource) applyFileMarks(filesToProcess []*files.File) error {
	for _, mark := range s.opts.fileMarks {
		pieces := strings.SplitN(mark, ":", 2)
		if len(pieces) != 2 {
			return fmt.Errorf("Expected file mark '%s' to be in format path:key=value", mark)
		}

		path := pieces[0]

		kv := strings.SplitN(pieces[1], "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("Expected file mark '%s' key-value portion to be in format key=value", mark)
		}

		var found bool

		for i, file := range filesToProcess {
			if file.OriginalRelativePath() == path {
				switch kv[0] {
				case "path":
					file.MarkRelativePath(kv[1])

				case "exclude":
					switch kv[1] {
					case "true":
						filesToProcess = append(filesToProcess[:i], filesToProcess[i+1:]...)
					default:
						return fmt.Errorf("Unknown value in file mark '%s'", mark)
					}

				case "type":
					switch kv[1] {
					case "yaml-template": // yaml template processing
						file.MarkType(files.TypeYAML)
						file.MarkTemplate(true)
					case "yaml-plain": // no template processing
						file.MarkType(files.TypeYAML)
						file.MarkTemplate(false)
					case "text-template":
						file.MarkType(files.TypeText)
						file.MarkTemplate(true)
					case "text-plain":
						file.MarkType(files.TypeText)
						file.MarkTemplate(false)
					case "starlark":
						file.MarkType(files.TypeStarlark)
						file.MarkTemplate(false)
					case "data":
						file.MarkType(files.TypeUnknown)
						file.MarkTemplate(false)
					default:
						return fmt.Errorf("Unknown value in file mark '%s'", mark)
					}

				case "output":
					// choose this template for output

				default:
					return fmt.Errorf("Unknown key in file mark '%s'", mark)
				}

				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("Expected file mark '%s' to match one file by path, but did not", mark)
		}
	}

	return nil
}
