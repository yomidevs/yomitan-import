/*
 * Copyright (c) 2016 Alex Yatskov <alex@foosoft.net>
 * Author: Alex Yatskov <alex@foosoft.net>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"os"
	"strconv"

	"github.com/FooSoft/jmdict"
)

const kanjidicRevision = "kanjidic1"

func kanjidicExtractKanji(entry jmdict.KanjidicCharacter, language string) *dbKanji {
	if entry.ReadingMeaning == nil {
		return nil
	}

	kanji := dbKanji{
		Character: entry.Literal,
		Indices:   make(map[string]string),
		Stats:     make(map[string]string),
	}

	for _, m := range entry.ReadingMeaning.Meanings {
		if m.Language == nil && language == "" || m.Language != nil && language == *m.Language {
			kanji.Meanings = append(kanji.Meanings, m.Meaning)
		}
	}

	if len(kanji.Meanings) == 0 {
		return nil
	}

	for _, number := range entry.DictionaryNumbers {
		kanji.Indices[number.Type] = number.Value
	}

	if frequency := entry.Misc.Frequency; frequency != nil {
		kanji.Stats["Frequency"] = *frequency
	}

	if level := entry.Misc.JlptLevel; level != nil {
		kanji.Stats["JLPT Level"] = *level
	}

	if counts := entry.Misc.StrokeCounts; len(counts) > 0 {
		kanji.Stats["Strokes"] = counts[0]
	}

	if grade := entry.Misc.Grade; grade != nil {
		kanji.Stats["Grade"] = *grade
		if gradeInt, err := strconv.Atoi(*grade); err == nil {
			if gradeInt >= 1 && gradeInt <= 8 {
				kanji.addTags("jouyou")
			} else if gradeInt >= 9 && gradeInt <= 10 {
				kanji.addTags("jinmeiyou")
			}
		}
	}

	for _, r := range entry.ReadingMeaning.Readings {
		switch r.Type {
		case "ja_on":
			kanji.Onyomi = append(kanji.Onyomi, r.Value)
		case "ja_kun":
			kanji.Kunyomi = append(kanji.Kunyomi, r.Value)
		}
	}

	return &kanji
}

func kanjidicExportDb(inputPath, outputPath, language, title string, stride int, pretty bool) error {
	reader, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	dict, err := jmdict.LoadKanjidic(reader)
	if err != nil {
		return err
	}

	var langTag string
	switch language {
	case "french":
		langTag = "fr"
	case "spanish":
		langTag = "es"
	case "portuguese":
		langTag = "pt"
	}

	var kanji dbKanjiList
	for _, entry := range dict.Characters {
		kanjiCurr := kanjidicExtractKanji(entry, langTag)
		if kanjiCurr != nil {
			kanji = append(kanji, *kanjiCurr)
		}
	}

	if title == "" {
		title = "KANJIDIC2"
	}

	tags := dbTagList{
		dbTag{Name: "jouyou", Notes: "included in list of regular-use characters", Category: "frequent", Order: -5},
		dbTag{Name: "jinmeiyou", Notes: "included in list of characters for use in personal names", Category: "frequent", Order: -5},
	}

	recordData := map[string]dbRecordList{
		"kanji": kanji.crush(),
		"tag":   tags.crush(),
	}

	return writeDb(
		outputPath,
		title,
		kanjidicRevision,
		recordData,
		stride,
		pretty,
	)
}
