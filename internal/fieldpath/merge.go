package fieldpath

import (
	"reflect"

	"github.com/imdario/mergo"
	"github.com/krateoplatformops/patch-provider/apis/patch/v1alpha1"
	"github.com/pkg/errors"
)

const (
	errInvalidMerge = "failed to merge values"
)

// MergeValue of the receiver p at the specified field path with the supplied
// value according to supplied merge options
func (p *Paved) MergeValue(path string, value any, mo *v1alpha1.MergeOptions) error {
	dst, err := p.GetValue(path)
	if IsNotFound(err) || mo == nil {
		dst = nil
	} else if err != nil {
		return err
	}

	dst, err = merge(dst, value, mo)
	if err != nil {
		return err
	}

	return p.SetValue(path, dst)
}

// merges the given src onto the given dst.
// dst and src must have the same type.
// If a nil merge options is supplied, the default behavior is MergeOptions'
// default behavior. If dst or src is nil, src is returned
// (i.e., dst replaced by src).
func merge(dst, src any, mergeOptions *v1alpha1.MergeOptions) (any, error) {
	// because we are merging values of a field, which can be a slice, and
	// because mergo currently supports merging only maps or structs,
	// we wrap the argument to be passed to mergo.Merge in a map.
	const keyArg = "arg"
	argWrap := func(arg any) map[string]any {
		return map[string]any{
			keyArg: arg,
		}
	}

	if dst == nil || src == nil {
		return src, nil // no merge, replace
	}

	// but, by default, do not append duplicate slice items if MergeOptions.AppendSlice is set
	if mergeOptions.IsAppendSlice() {
		src = removeSourceDuplicates(dst, src)
	}

	mDst := argWrap(dst)
	// use merge semantics with the configured merge options to obtain the target dst value
	if err := mergo.Merge(&mDst, argWrap(src), mergeOptions.MergoConfiguration()...); err != nil {
		return nil, errors.Wrap(err, errInvalidMerge)
	}
	return mDst[keyArg], nil
}

func removeSourceDuplicates(dst, src any) any {
	sliceDst, sliceSrc := reflect.ValueOf(dst), reflect.ValueOf(src)
	if sliceDst.Kind() == reflect.Ptr {
		sliceDst = sliceDst.Elem()
	}
	if sliceSrc.Kind() == reflect.Ptr {
		sliceSrc = sliceSrc.Elem()
	}
	if sliceDst.Kind() != reflect.Slice || sliceSrc.Kind() != reflect.Slice {
		return src
	}

	result := reflect.New(sliceSrc.Type()).Elem() // we will not modify src
	for i := 0; i < sliceSrc.Len(); i++ {
		itemSrc := sliceSrc.Index(i)
		found := false
		for j := 0; j < sliceDst.Len() && !found; j++ {
			// if src item is found in the dst array
			if reflect.DeepEqual(itemSrc.Interface(), sliceDst.Index(j).Interface()) {
				found = true
			}
		}
		if !found {
			// then put src item into result
			result = reflect.Append(result, itemSrc)
		}
	}
	return result.Interface()
}
