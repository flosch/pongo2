package pongo2

// tombstone marks a key as deleted within a child Scope. It masks a value
// inherited from a parent Scope without mutating the parent, so the deletion is
// only visible to the child and its descendants.
var tombstone = &struct{}{}

// Scope is a chain of variable layers representing nested template scopes (for
// example the body of a {% for %} or {% with %} block). Each child layers its
// own variables over its parent without copying, so lookups walk outward from
// the innermost scope to the outermost. Writes only ever touch the innermost
// (own) layer; a parent scope is never modified by its children.
//
// Scope is the type of ExecutionContext.Private. It is a lightweight handle onto
// its owning ExecutionContext and the chain of parents — there is no separate
// per-scope allocation. Access it through the Get/Set/Delete/Has/Range methods
// rather than indexing a map.
type Scope struct {
	ec *ExecutionContext
}

// Get returns the value bound to key, searching this scope and then its parents.
// The second return value reports whether the key was found. A key masked by
// Delete in a closer scope is reported as not found.
func (s Scope) Get(key string) (any, bool) {
	for c := s.ec; c != nil; c = c.parent {
		if v, ok := c.privateVars[key]; ok {
			if v == tombstone {
				return nil, false
			}
			return v, true
		}
	}
	return nil, false
}

// Has reports whether key is bound in this scope or any parent.
func (s Scope) Has(key string) bool {
	_, ok := s.Get(key)
	return ok
}

// Set binds key to value in this scope's own layer, shadowing any binding
// inherited from a parent. Parent scopes are never modified. The own layer is
// allocated lazily on first write.
func (s Scope) Set(key string, value any) {
	if s.ec.privateVars == nil {
		s.ec.privateVars = make(Context)
	}
	s.ec.privateVars[key] = value
}

// Delete removes key from this scope's view. If the key lives only in this
// scope's own layer it is removed outright; if it is inherited from a parent it
// is masked with a tombstone so the parent remains untouched.
func (s Scope) Delete(key string) {
	delete(s.ec.privateVars, key)
	if s.ec.parent != nil {
		if _, ok := (Scope{ec: s.ec.parent}).Get(key); ok {
			if s.ec.privateVars == nil {
				s.ec.privateVars = make(Context)
			}
			s.ec.privateVars[key] = tombstone
		}
	}
}

// Range calls fn for each key visible in this scope, with closer scopes
// shadowing parents and tombstoned keys skipped. Iteration stops early if fn
// returns false.
func (s Scope) Range(fn func(key string, value any) bool) {
	for key, value := range s.flatten() {
		if !fn(key, value) {
			return
		}
	}
}

// flatten collapses the scope chain into a single map, with closer scopes
// overriding parents and tombstoned keys removed. Used by tags that need a
// full snapshot of the private context (for example {% include %}).
func (s Scope) flatten() Context {
	return flattenScope(s.ec)
}

func flattenScope(ec *ExecutionContext) Context {
	if ec == nil {
		return make(Context)
	}
	out := flattenScope(ec.parent)
	for key, value := range ec.privateVars {
		if value == tombstone {
			delete(out, key)
			continue
		}
		out[key] = value
	}
	return out
}

// PublicContext is a read-only view of the data available to a template: the
// user-supplied context overlaying the template set's globals. It is read-only
// by design — a template must not mutate caller-provided data — so it exposes
// no setter. Globals are reached only through Get/Has/Range, so callers cannot
// bypass them by indexing a map directly.
type PublicContext struct {
	vars    Context // user-supplied context (used directly, never copied)
	globals Context // template set globals
}

// Get returns the value bound to key, checking the user context first and then
// the set globals. The second return value reports whether key was found.
func (p PublicContext) Get(key string) (any, bool) {
	if v, ok := p.vars[key]; ok {
		return v, true
	}
	if v, ok := p.globals[key]; ok {
		return v, true
	}
	return nil, false
}

// Has reports whether key is available in the user context or globals.
func (p PublicContext) Has(key string) bool {
	_, ok := p.Get(key)
	return ok
}

// Range calls fn for each key available, with user-context entries shadowing
// globals of the same name. Iteration stops early if fn returns false.
func (p PublicContext) Range(fn func(key string, value any) bool) {
	for key, value := range p.globals {
		if _, shadowed := p.vars[key]; shadowed {
			continue
		}
		if !fn(key, value) {
			return
		}
	}
	for key, value := range p.vars {
		if !fn(key, value) {
			return
		}
	}
}
