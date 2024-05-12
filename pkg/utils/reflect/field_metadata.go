// Copyright (c) 2024 David Ouellette.
//
// All rights reserved.
//
// This software and its documentation are proprietary information of David Ouellette.
// No part of this software or its documentation may be copied, transferred, reproduced,
// distributed, modified, or disclosed without the prior written permission of David Ouellette.
//
// Unauthorized use of this software is strictly prohibited and may be subject to civil and
// criminal penalties.
//
// By using this software, you agree to abide by the terms specified herein.

package reflect

import (
	"fmt"
	"reflect"
	"regexp"
	"sync"
)

var (
	// fieldsToMetadataMemo is used to cache the result of the FieldsToMetadata function.
	fieldsToMetadataMemo = sync.Map{}

	// tagMatchRegex matches all tag entries on a struct field.
	tagMatchRegex = regexp.MustCompile(`(\w+):"([^"]*)"`)
)

// FieldMetadata is the metadata extracted from struct fields.
type FieldMetadata struct {
	Type      reflect.Type
	Tags      map[string]string
	Anonymous []string
}

// FieldsToMetadata returns a map of a structs field names to their respective metadata.
// The returned map should not be written to under any circumstances since it can be shared among many threads.
func FieldsToMetadata[T any]() map[string]*FieldMetadata {
	reflectType := reflect.TypeOf(*new(T))
	if memoData, ok := fieldsToMetadataMemo.Load(reflectType); ok {
		return memoData.(map[string]*FieldMetadata)
	}

	fieldsToMetadata := make(map[string]*FieldMetadata)
	processType(reflectType, fieldsToMetadata, make([]string, 0))

	fieldsToMetadataMemo.Store(reflectType, fieldsToMetadata)
	return fieldsToMetadata
}

// processType takes a struct type, lists all of its fields, and builds the metadata for it.
// If the struct contains an embedded anonymous struct, it appends its name to the anonymous name chain.
// If a field name is not unique, a panic occurs. This includes field names of the anonymous structs.
func processType(reflectType reflect.Type, fieldsToMetadata map[string]*FieldMetadata, anonymousChain []string) {
	if reflectType.Kind() != reflect.Struct {
		panic("type must be a struct")
	}

	for fieldIndex := 0; fieldIndex < reflectType.NumField(); fieldIndex++ {
		field := reflectType.Field(fieldIndex)

		anonymousChainCopy := make([]string, len(anonymousChain))
		copy(anonymousChainCopy, anonymousChain)

		if field.Anonymous {
			anonymousChainCopy = append(anonymousChainCopy, field.Name)
			processType(field.Type, fieldsToMetadata, anonymousChainCopy)
			continue
		}

		if _, alreadyHasFieldName := fieldsToMetadata[field.Name]; alreadyHasFieldName {
			panic(fmt.Sprintf("field %s is ambiguous", field.Name))
		}

		metadata := &FieldMetadata{}
		metadata.Type = field.Type
		metadata.Tags = make(map[string]string)
		metadata.Anonymous = anonymousChainCopy

		if len(string(field.Tag)) != 0 {
			matches := tagMatchRegex.FindAllStringSubmatch(string(field.Tag), -1)
			for _, match := range matches {
				tagKey := match[1]
				tagValue := match[2]
				metadata.Tags[tagKey] = tagValue
			}
		}

		fieldsToMetadata[field.Name] = metadata
	}
}
