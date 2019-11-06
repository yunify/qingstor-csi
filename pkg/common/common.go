/*
Copyright (C) 2018 Yunify, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this work except in compliance with the License.
You may obtain a copy of the License in the LICENSE file, or at:

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"context"
	"fmt"
	"hash/fnv"
)

// ContextWithHash return context with hash value
func ContextWithHash(ctx context.Context, hash string) context.Context {
	return context.WithValue(ctx, Hash, hash)
}

// GetContextHash return hash of context
func GetContextHash(ctx context.Context) string {
	val := ctx.Value(Hash)
	if v, ok := val.(string); ok {
		return v
	}
	return ""
}

// GenerateRandIdSuffix generates a random resource id.
func GenerateHashInEightBytes(input string) string {
	h := fnv.New32a()
	_, err := h.Write([]byte(input))
	if err != nil {
		return "Error Hash!!!!!"
	}
	return fmt.Sprintf("%.8x", h.Sum32())
}
