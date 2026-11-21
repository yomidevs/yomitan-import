package yomitan

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportEdrdgXML(t *testing.T) {
	const kanji = `<character><literal>山</literal><reading_meaning><rmgroup>
<reading r_type="ja_on">サン</reading><meaning>mountain</meaning>
</rmgroup></reading_meaning></character></kanjidic2>`
	const name = `<entry><ent_seq>5000014</ent_seq><k_ele><keb>ＡＮＣ</keb></k_ele>
<r_ele><reb>エーエヌシー</reb></r_ele>
<trans><name_type>&TYPE;</name_type><trans_det>African National Congress</trans_det></trans>
</entry>`
	const nameLegacy = `<!DOCTYPE JMnedict [
<!ENTITY organization "organization name">
]>
<JMnedict>` + name + `
<entry><ent_seq>9999990</ent_seq><r_ele><reb>ジェイエムエヌイーディクト</reb></r_ele>
<trans><trans_det>Creation Date: 2026-01-02</trans_det></trans></entry></JMnedict>`
	const nameNG = `<!DOCTYPE JMnedict [
<!ENTITY org "organization name (alias)">
<!ENTITY organization "organization name">
]>
<!-- JMnedict created: 2026-01-02 -->
<JMnedict created="2026-01-02" version="1.09">` + name + `</JMnedict>`

	for _, tt := range []struct {
		filename, document, bank, term, title, revision string
	}{
		{"kanjidic2.xml", `<kanjidic2><header><database_version>2026-002</database_version></header>` + kanji,
			"kanji_bank_1.json", `"mountain"`, "KANJIDIC2 [2026-002]", "kanjidic2.2026-002"},
		{"kanjidic2_NG.xml", `<kanjidic2 created="2026-01-02" version="1.6">` + kanji,
			"kanji_bank_1.json", `"mountain"`, "KANJIDIC2 [2026-002]", "kanjidic2.2026-002"},
		{"JMnedict.xml", strings.ReplaceAll(nameLegacy, "TYPE", "organization"),
			"term_bank_1.json", `"organization"`, "JMnedict [2026-01-02]", "JMnedict.2026-01-02"},
		{"renamed.xml", strings.ReplaceAll(nameNG, "TYPE", "org"),
			"term_bank_1.json", `"organization"`, "JMnedict [2026-01-02]", "JMnedict.2026-01-02"},
	} {
		t.Run(tt.filename, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, tt.filename)
			output := filepath.Join(dir, "dict.zip")
			if err := os.WriteFile(input, []byte("<?xml version=\"1.0\"?>\n"+tt.document), 0600); err != nil {
				t.Fatal(err)
			}
			if err := ExportDb(input, output, DefaultFormat, DefaultLanguage, DefaultTitle, DefaultStride, DefaultPretty); err != nil {
				t.Fatal(err)
			}
			z, err := zip.OpenReader(output)
			if err != nil {
				t.Fatal(err)
			}
			defer z.Close()

			var index dbIndex
			readDictionaryBank(t, &z.Reader, "index.json", &index)
			if index.Title != tt.title || index.Revision != tt.revision {
				t.Errorf("got title %q revision %q, want %q and %q", index.Title, index.Revision, tt.title, tt.revision)
			}
			var content json.RawMessage
			readDictionaryBank(t, &z.Reader, tt.bank, &content)
			if !strings.Contains(string(content), tt.term) {
				t.Errorf("missing %s in %s: %s", tt.term, tt.bank, content)
			}
			if tt.bank == "term_bank_1.json" {
				var tags json.RawMessage
				readDictionaryBank(t, &z.Reader, "tag_bank_1.json", &tags)
				if !strings.Contains(string(tags), `"organization name"`) || strings.Contains(string(tags), `"org"`) {
					t.Errorf("organization tag not normalized: %s", tags)
				}
			}
		})
	}
}

func readDictionaryBank(t *testing.T, z *zip.Reader, name string, value any) {
	t.Helper()
	f, err := z.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(value); err != nil {
		t.Fatal(err)
	}
}

func TestDetectFormatKeepsExistingFormats(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CATALOGS"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	for name, expected := range map[string]string{
		dir: "epwing", filepath.Join(dir, "JMdict_e_examp"): "edict",
		filepath.Join(dir, "dict.sqlite"): "rikai", filepath.Join(dir, "dict.kanjifreq"): "kanjifreq",
		filepath.Join(dir, "dict.termfreq"): "termfreq",
	} {
		if got, err := detectFormat(name); err != nil || got != expected {
			t.Errorf("detectFormat(%q) = %q, %v; want %q", name, got, err, expected)
		}
	}
	input := filepath.Join(dir, "unknown.xml")
	for _, document := range []string{"", "<kanjidic2", "<unknown/>"} {
		if err := os.WriteFile(input, []byte(document), 0600); err != nil {
			t.Fatal(err)
		}
		if got, err := detectFormat(input); err == nil {
			t.Errorf("detectFormat accepted %q as %q", document, got)
		}
	}
}

func TestRejectInvalidEdrdgXMLRevision(t *testing.T) {
	for _, document := range []string{
		`<kanjidic2 created="2026-02-30" version="1.6"/>`,
		`<JMnedict version="1.09"/>`,
	} {
		dir := t.TempDir()
		input := filepath.Join(dir, "input.xml")
		if err := os.WriteFile(input, []byte(document), 0600); err != nil {
			t.Fatal(err)
		}
		err := ExportDb(input, filepath.Join(dir, "dict.zip"), DefaultFormat, DefaultLanguage, DefaultTitle, DefaultStride, DefaultPretty)
		if err == nil || !strings.Contains(err.Error(), "creation date") {
			t.Errorf("ExportDb(%s) produced %v, want a creation date error", document, err)
		}
	}
}
