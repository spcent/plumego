package pubsub

import "reflect"

func cloneMessage(msg Message) Message {
	if msg.Meta != nil {
		meta := make(map[string]string, len(msg.Meta))
		for k, v := range msg.Meta {
			meta[k] = v
		}
		msg.Meta = meta
	}
	msg.Data = cloneData(msg.Data)
	return msg
}

func cloneData(data any) any {
	if data == nil {
		return nil
	}
	if copier, ok := data.(interface{ DeepCopy() any }); ok {
		return copier.DeepCopy()
	}
	if copier, ok := data.(interface{ Clone() any }); ok {
		return copier.Clone()
	}

	cloned, ok := cloneValue(reflect.ValueOf(data))
	if ok && cloned.IsValid() {
		return cloned.Interface()
	}
	return data
}

func cloneValue(v reflect.Value) (reflect.Value, bool) {
	if !v.IsValid() {
		return v, true
	}

	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return v, true
		}
		cloned, _ := cloneValue(v.Elem())
		if !cloned.IsValid() {
			return v, true
		}
		out := reflect.New(v.Type()).Elem()
		if cloned.Type().AssignableTo(v.Type()) {
			out.Set(cloned)
			return out, true
		}
		if cloned.Type().ConvertibleTo(v.Type()) {
			out.Set(cloned.Convert(v.Type()))
			return out, true
		}
		out.Set(v.Elem())
		return out, true
	case reflect.Ptr:
		if v.IsNil() {
			return v, true
		}
		cloned, _ := cloneValue(v.Elem())
		if !cloned.IsValid() {
			return v, true
		}
		out := reflect.New(v.Elem().Type())
		if cloned.Type().AssignableTo(v.Elem().Type()) {
			out.Elem().Set(cloned)
			return out, true
		}
		if cloned.Type().ConvertibleTo(v.Elem().Type()) {
			out.Elem().Set(cloned.Convert(v.Elem().Type()))
			return out, true
		}
		out.Elem().Set(v.Elem())
		return out, true
	case reflect.Map:
		if v.IsNil() {
			return v, true
		}
		out := reflect.MakeMapWithSize(v.Type(), v.Len())
		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)
			clonedValue, _ := cloneValue(value)
			if !clonedValue.IsValid() {
				clonedValue = value
			}
			if clonedValue.Type() != value.Type() {
				if clonedValue.Type().AssignableTo(value.Type()) {
					clonedValue = clonedValue.Convert(value.Type())
				} else if clonedValue.Type().ConvertibleTo(value.Type()) {
					clonedValue = clonedValue.Convert(value.Type())
				} else {
					clonedValue = value
				}
			}
			out.SetMapIndex(key, clonedValue)
		}
		return out, true
	case reflect.Slice:
		if v.IsNil() {
			return v, true
		}
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		if v.Type().Elem().Kind() == reflect.Uint8 {
			reflect.Copy(out, v)
			return out, true
		}
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			clonedElem, _ := cloneValue(elem)
			if !clonedElem.IsValid() {
				clonedElem = elem
			}
			if clonedElem.Type() != elem.Type() {
				if clonedElem.Type().AssignableTo(elem.Type()) {
					clonedElem = clonedElem.Convert(elem.Type())
				} else if clonedElem.Type().ConvertibleTo(elem.Type()) {
					clonedElem = clonedElem.Convert(elem.Type())
				} else {
					clonedElem = elem
				}
			}
			out.Index(i).Set(clonedElem)
		}
		return out, true
	case reflect.Array:
		out := reflect.New(v.Type()).Elem()
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			clonedElem, _ := cloneValue(elem)
			if !clonedElem.IsValid() {
				clonedElem = elem
			}
			if clonedElem.Type() != elem.Type() {
				if clonedElem.Type().AssignableTo(elem.Type()) {
					clonedElem = clonedElem.Convert(elem.Type())
				} else if clonedElem.Type().ConvertibleTo(elem.Type()) {
					clonedElem = clonedElem.Convert(elem.Type())
				} else {
					clonedElem = elem
				}
			}
			out.Index(i).Set(clonedElem)
		}
		return out, true
	default:
		return v, true
	}
}
