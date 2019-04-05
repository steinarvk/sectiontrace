package sectiontrace

import (
	"context"
	"fmt"
)

type parentNodeKey struct{}
type ancestorNodeKey struct{}
type remoteParentNodeKey struct{}
type remoteParentScopeKey struct{}
type remoteAncestorNodeKey struct{}
type remoteAncestorScopeKey struct{}

var ParentNodeContextKey = parentNodeKey{}
var AncestorNodeContextKey = ancestorNodeKey{}
var RemoteParentNodeContextKey = remoteParentNodeKey{}
var RemoteParentScopeContextKey = remoteParentScopeKey{}
var RemoteAncestorNodeContextKey = remoteAncestorNodeKey{}
var RemoteAncestorScopeContextKey = remoteAncestorScopeKey{}

type NodeAndScope struct {
	Scope string `json:"scope"`
	ID    int32  `json:"id"`
}

type RemoteInfo struct {
	Parent   NodeAndScope `json:"parent"`
	Ancestor NodeAndScope `json:"ancestor"`
}

func RemoteInfoFromContext(ctx context.Context) (*RemoteInfo, error) {
	var rv RemoteInfo

	if v := ctx.Value(RemoteParentNodeContextKey); v != nil {
		unpacked, ok := v.(int32)
		if !ok {
			return nil, fmt.Errorf("Invalid value for RemoteParentNodeContextKey: %v", v)
		}
		rv.Parent.ID = unpacked
	}

	if v := ctx.Value(RemoteAncestorNodeContextKey); v != nil {
		unpacked, ok := v.(int32)
		if !ok {
			return nil, fmt.Errorf("Invalid value for RemoteAncestorNodeContextKey: %v", v)
		}
		rv.Ancestor.ID = unpacked
	}

	if v := ctx.Value(RemoteParentScopeContextKey); v != nil {
		unpacked, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("Invalid value for RemoteParentScopeContextKey: %v", v)
		}
		rv.Parent.Scope = unpacked
	}

	if v := ctx.Value(RemoteAncestorScopeContextKey); v != nil {
		unpacked, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("Invalid value for RemoteAncestorScopeContextKey: %v", v)
		}
		rv.Ancestor.Scope = unpacked
	}

	emptyRemoteInfo := RemoteInfo{}

	if rv == emptyRemoteInfo {
		return nil, nil
	}

	if rv.Ancestor.Scope == "" {
		return nil, fmt.Errorf("Missing ancestor scope")
	}

	if rv.Parent.Scope == "" {
		return nil, fmt.Errorf("Missing parent scope")
	}

	if rv.Parent.ID == 0 {
		return nil, fmt.Errorf("Missing parent ID")
	}

	if rv.Ancestor.ID == 0 {
		return nil, fmt.Errorf("Missing ancestor ID")
	}

	return &rv, nil
}

func ContextWithRemoteInfo(ctx context.Context, info *RemoteInfo) context.Context {
	if info == nil {
		return ctx
	}
	ctx = context.WithValue(ctx, RemoteParentNodeContextKey, info.Parent.ID)
	ctx = context.WithValue(ctx, RemoteParentScopeContextKey, info.Parent.Scope)
	ctx = context.WithValue(ctx, RemoteAncestorNodeContextKey, info.Ancestor.ID)
	ctx = context.WithValue(ctx, RemoteAncestorScopeContextKey, info.Ancestor.Scope)
	return ctx
}
