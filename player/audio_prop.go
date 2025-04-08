package player

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/dhowden/tag"
)

// audioProperties represents metadata about an audio file.
type audioProperties struct {
	FileName string
	Title    string
	Artist   string
	Album    string
	Genre    string
	Year     int
}

// Display prints the audio properties to the console.
func (p *audioProperties) Display() {
	fmt.Println(p.FileName)
	fmt.Println(strings.Repeat("-", p.delimiterLen()))
	fmt.Printf("Title  | %s\n", p.Title)
	fmt.Printf("Artist | %s\n", p.Artist)
	fmt.Printf("Album  | %s\n", p.Album)
	fmt.Printf("Genre  | %s\n", p.Genre)
	fmt.Printf("Year   | %d\n", p.Year)
}

// delimiterLen returns the length of the horizontal delimiter used in the Display method,
// which is calculated as the maximum length of the string fields in the audioProperties structure.
func (p *audioProperties) delimiterLen() int {
	maxLen := 0
	rv := reflect.ValueOf(p).Elem()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if field.Kind() == reflect.String {
			fl := len([]rune(field.String()))
			if i != 0 {
				fl += 9 // add offset for fieldname plus separator
			}
			if fl > maxLen {
				maxLen = fl
			}
		}
	}

	return maxLen
}

func NewAudioProperties(filePath string) (*audioProperties, error) {
	prop := audioProperties{
		Title:  "-",
		Artist: "-",
		Album:  "-",
		Genre:  "-",
	}

	f, err := os.Open(filePath)
	if err != nil {
		return &prop, err
	}

	prop.FileName = filepath.Base(f.Name())

	m, err := tag.ReadFrom(f)
	if err != nil {
		return &prop, err
	}

	prop.Title = m.Title()
	prop.Artist = m.Artist()
	prop.Album = m.Album()
	prop.Genre = m.Genre()
	prop.Year = m.Year()

	return &prop, err
}
