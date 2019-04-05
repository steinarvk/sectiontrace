package sectiontrace

import (
	"context"
	"fmt"
)

const ArgParent = "p"
const ArgAncestor = "a"
const ArgRemoteParent = "rp"
const ArgRemoteParentScope = "rps"
const ArgRemoteAncestor = "ra"
const ArgRemoteAncestorScope = "ras"
const ArgOK = "ok"

func setArgsFromContext(ctx context.Context, args map[string]interface{}) error {
	if v := ctx.Value(ParentNodeContextKey); v != nil {
		unpacked, ok := v.(int32)
		if !ok {
			return fmt.Errorf("Invalid value for ParentNodeContextKey: %v", v)
		}
		args[ArgParent] = unpacked
	}

	if v := ctx.Value(AncestorNodeContextKey); v != nil {
		unpacked, ok := v.(int32)
		if !ok {
			return fmt.Errorf("Invalid value for AncestorNodeContextKey: %v", v)
		}
		args[ArgAncestor] = unpacked
	}

	info, err := RemoteInfoFromContext(ctx)
	if err != nil {
		return err
	}
	if info != nil {
		args[ArgRemoteParent] = info.Parent.ID
		args[ArgRemoteAncestor] = info.Ancestor.ID
		args[ArgRemoteParentScope] = info.Parent.Scope
		args[ArgRemoteAncestorScope] = info.Ancestor.Scope
	}

	return nil
}
