package tmf

// Resources is the top-level resource list of a 3MF model part. It holds
// Objects, BaseMaterials, and extension-supplied resources (color groups,
// texture groups, slice stacks, etc.) accessed via Extension/SetExtension.
type Resources struct {
	objects       []*Object
	baseMaterials []*BaseMaterials

	// extensionResources holds extension-supplied resource collections keyed
	// by namespace URI. The value is the extension package's own resource
	// type (typically a slice of resources).
	extensionResources map[string]any
}

// Objects returns all Object resources in this list.
func (r *Resources) Objects() []*Object { return r.objects }

// AppendObject appends an object resource.
func (r *Resources) AppendObject(o *Object) { r.objects = append(r.objects, o) }

// FindObject returns the Object with the given id, or nil if none matches.
// O(n) — for tight loops, build a map.
func (r *Resources) FindObject(id uint32) *Object {
	for _, o := range r.objects {
		if o.id == id {
			return o
		}
	}
	return nil
}

// BaseMaterials returns the BaseMaterials resources in this list.
func (r *Resources) BaseMaterials() []*BaseMaterials { return r.baseMaterials }

// AppendBaseMaterials appends a base-materials resource.
func (r *Resources) AppendBaseMaterials(b *BaseMaterials) {
	r.baseMaterials = append(r.baseMaterials, b)
}

// FindBaseMaterials returns the BaseMaterials with the given id, or nil.
func (r *Resources) FindBaseMaterials(id uint32) *BaseMaterials {
	for _, b := range r.baseMaterials {
		if b.id == id {
			return b
		}
	}
	return nil
}

// ExtensionResources returns the extension-specific resource value for
// namespace ns, or nil.
func (r *Resources) ExtensionResources(ns string) any {
	if r.extensionResources == nil {
		return nil
	}
	return r.extensionResources[ns]
}

// SetExtensionResources attaches an extension-specific resource value.
// Passing nil removes it.
func (r *Resources) SetExtensionResources(ns string, v any) {
	if v == nil {
		delete(r.extensionResources, ns)
		return
	}
	if r.extensionResources == nil {
		r.extensionResources = make(map[string]any)
	}
	r.extensionResources[ns] = v
}
