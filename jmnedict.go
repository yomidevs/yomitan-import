package yomitan

import (
	"fmt"
	"os"
	"regexp"
	"time"

	jmdict "github.com/yomidevs/jmdict-go"
)

func jmnedictPublicationDate(dictionary jmdict.Jmnedict) string {
	unknownDate := "unknown"
	idx := len(dictionary.Entries) - 1
	if len(dictionary.Entries) == 0 {
		return unknownDate
	} else if len(dictionary.Entries[idx].Translations) == 0 {
		return unknownDate
	} else if len(dictionary.Entries[idx].Translations[0].Translations) == 0 {
		return unknownDate
	}
	dateGloss := dictionary.Entries[idx].Translations[0].Translations[0]
	r := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	date := r.FindString(dateGloss)
	if date != "" {
		return date
	} else {
		return unknownDate
	}
}

func jmnedictSenseTerm(headword headword, seq sequence, sense jmdict.JmnedictTranslation, senseNumber int) dbTerm {
	term := dbTerm{
		Expression: headword.Expression,
		Reading:    headword.Reading,
		Sequence:   seq,
	}
	for _, gloss := range sense.Translations {
		term.Glossary = append(term.Glossary, gloss)
	}
	term.addDefinitionTags(sense.NameTypes...)
	term.Score = calculateTermScore(senseNumber, 0, headword)
	return term
}

func jmnedictTerms(headword headword, entry jmdict.JmnedictEntry, g genericTermInfo) []dbTerm {
	terms := []dbTerm{}
	for idx, sense := range entry.Translations {
		if g.IsGenericName(headword, sense.Translations) {
			g.AddGlosses(headword.Expression, sense.NameTypes, headword.Reading)
		} else {
			g.AddUsedSequence(entry.Sequence)
			senseTerm := jmnedictSenseTerm(headword, entry.Sequence, sense, idx+1)
			terms = append(terms, senseTerm)
		}
	}
	return terms
}

func jmnedictHeadwords(entry jmdict.JmnedictEntry) (headwords []headword) {
	// Note that JMnedict doesn't (currently) use priority tags,
	// frequency tags, or any sort of reading/kanji restrictions.
	for _, reading := range entry.Readings {
		for _, kanji := range entry.Kanji {
			h := headword{
				Expression: kanji.Expression,
				Reading:    reading.Reading,
			}
			h.Index = len(headwords)
			headwords = append(headwords, h)
		}
	}
	if len(entry.Kanji) == 0 {
		for _, reading := range entry.Readings {
			h := headword{
				Expression: reading.Reading,
				Reading:    reading.Reading,
			}
			h.Index = len(headwords)
			headwords = append(headwords, h)
		}
	}
	return headwords
}

func jmnedictExportDb(inputPath, outputPath, language, title string, stride int, pretty bool) error {
	reader, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	created, version, err := dictionaryXMLMetadata(reader, "JMnedict")
	if err != nil {
		return err
	}
	dictionary, entities, err := jmdict.LoadJmnedictNoTransform(reader)
	if err != nil {
		return err
	}

	jmnedictDate := jmnedictPublicationDate(dictionary)
	if created != "" {
		date, err := time.Parse("2006-01-02", created)
		if err != nil {
			return fmt.Errorf("invalid JMnedict creation date: %w", err)
		}
		jmnedictDate = date.Format("2006-01-02")
	} else if version != "" && jmnedictDate == "unknown" {
		return fmt.Errorf("JMnedict version %s is missing its creation date", version)
	}

	// NG calls the same name type "org". Keep existing tags in both term
	// banks and tag definitions, including for generic-name grouping.
	if notes, ok := entities["org"]; ok {
		if _, exists := entities["organization"]; !exists {
			entities["organization"] = notes
		}
		delete(entities, "org")
	}
	genericTermInfo := newGenericTermInfo()

	terms := dbTermList{}
	for _, entry := range dictionary.Entries {
		for _, sense := range entry.Translations {
			for i, nameType := range sense.NameTypes {
				if nameType == "org" {
					sense.NameTypes[i] = "organization"
				}
			}
		}
		headwords := jmnedictHeadwords(entry)
		for _, headword := range headwords {
			newTerms := jmnedictTerms(headword, entry, genericTermInfo)
			terms = append(terms, newTerms...)
		}
	}
	terms = append(terms, genericTermInfo.Terms()...)

	tags := dbTagList{}
	tags = append(tags, entityTags(entities)...)

	recordData := map[string]dbRecordList{
		"term": terms.crush(),
		"tag":  tags.crush(),
	}

	if title == "" {
		title = "JMnedict"
	}
	title = title + " [" + jmnedictDate + "]"

	index := dbIndex{
		Title:       title,
		Revision:    "JMnedict." + jmnedictDate,
		Sequenced:   true,
		Attribution: edrdgAttribution,
	}

	return writeDb(
		outputPath,
		index,
		recordData,
		stride,
		pretty,
	)
}
