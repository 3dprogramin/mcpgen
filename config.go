package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// kv is one key/value pair of a JSON object, value kept raw.
type kv struct {
	Key   string
	Value json.RawMessage
}

// orderedObject is a JSON object that remembers its key order, so a server
// config round-trips as written instead of being alphabetised by a map.
type orderedObject []kv

func parseOrderedObject(data []byte) (orderedObject, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("expected a JSON object")
	}
	var obj orderedObject
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("expected string key")
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, err
		}
		obj = append(obj, kv{Key: key, Value: val})
	}
	return obj, nil
}

// MarshalJSON re-emits the object in its original key order. The result is not
// pretty-printed; callers run it through json.Indent for that.
func (o orderedObject) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, pair := range o {
		if i > 0 {
			b.WriteByte(',')
		}
		key, err := json.Marshal(pair.Key)
		if err != nil {
			return nil, err
		}
		b.Write(key)
		b.WriteByte(':')
		b.Write(pair.Value)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

// get returns the value for a key and whether it was present.
func (o orderedObject) get(key string) (json.RawMessage, bool) {
	for _, pair := range o {
		if pair.Key == key {
			return pair.Value, true
		}
	}
	return nil, false
}

// set replaces the value for an existing key (no-op if absent).
func (o orderedObject) set(key string, val json.RawMessage) {
	for i := range o {
		if o[i].Key == key {
			o[i].Value = val
			return
		}
	}
}

// serverArgs returns the "args" array of a server config, if it has one.
func serverArgs(config json.RawMessage) ([]string, bool) {
	obj, err := parseOrderedObject(config)
	if err != nil {
		return nil, false
	}
	raw, ok := obj.get("args")
	if !ok {
		return nil, false
	}
	var args []string
	if json.Unmarshal(raw, &args) != nil {
		return nil, false
	}
	return args, true
}

// applyArgs merges extra args into a server config's "args" array, preserving
// key order. A "--flag=value" override replaces an existing "--flag=..." entry;
// anything else is appended if not already present.
func applyArgs(config json.RawMessage, extra []string) (json.RawMessage, error) {
	if len(extra) == 0 {
		return config, nil
	}
	obj, err := parseOrderedObject(config)
	if err != nil {
		return nil, err
	}
	raw, ok := obj.get("args")
	if !ok {
		return nil, fmt.Errorf("this server has no \"args\" to customize")
	}
	var args []string
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, fmt.Errorf("server \"args\" is not a string array: %w", err)
	}

	args = mergeArgs(args, extra)

	newRaw, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	obj.set("args", newRaw)
	return json.Marshal(obj)
}

// mergeArgs applies override tokens to an existing arg list. A "--flag=value"
// override replaces a matching flag whether the existing arg uses the joined
// form ("--flag=old") or the split form ("--flag", "old"). Tokens that aren't
// "--flag=value" are appended unless already present.
func mergeArgs(existing, extra []string) []string {
	for _, e := range extra {
		eq := strings.IndexByte(e, '=')
		if eq <= 0 || !strings.HasPrefix(e, "-") {
			if !contains(existing, e) {
				existing = append(existing, e)
			}
			continue
		}

		flag := e[:eq]     // "--connectionString"
		prefix := e[:eq+1] // "--connectionString="
		value := e[eq+1:]  // "mongodb://..."
		replaced := false
		for j := 0; j < len(existing); j++ {
			switch {
			case strings.HasPrefix(existing[j], prefix): // joined form
				existing[j] = e
			case existing[j] == flag: // split form: value is the next token
				if j+1 < len(existing) {
					existing[j+1] = value
				} else {
					existing = append(existing, value)
				}
			default:
				continue
			}
			replaced = true
			break
		}
		if !replaced {
			existing = append(existing, e)
		}
	}
	return existing
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
