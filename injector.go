package digo

import (
	"errors"
	"reflect"
)

func init() {
	if depInjector == nil {
		depInjector = new(Injector)
	}
}

var depInjector *Injector

type Injector struct{}

func (this *Injector) resolve(node *NodeData, nodeMap map[string]*NodeData) (interface{}, error) {
	cp, err := this.newTypeOf(node.Type, node.IsPtr)
	if err != nil {
		return struct{}{}, errors.New("Error creating new Type -> " + err.Error())
	}

	for _, dependency := range node.Deps {

		err := this.assignValues(cp, dependency, nodeMap, node.IsPtr)
		if err != nil {
			return struct{}{}, errors.New("Error resolving dependencies -> " + err.Error())
		}

	}

	i := cp.Interface()

	if init, isInit := i.(Initializer); isInit {

		err := init.BeforeInject()
		if err != nil {
			return struct{}{}, nil
		}

	}

	return i, nil
}

func (this *Injector) newTypeOf(key string, isPtr bool) (reflect.Value, error) {
	t, err := TypeRegistry.Get(key)
	if err != nil {
		return reflect.Value{}, errors.New("Error getting the type from TypeRegistry -> " + err.Error())
	}

	if isPtr {
		return reflect.New(t), nil
	}

	return reflect.New(t).Elem(), nil
}

func (this *Injector) assignValues(cp reflect.Value, dependency *DepData, nodeMap map[string]*NodeData, isPtr bool) error {
	var depRoot *NodeData

	if depNode, exists := nodeMap[dependency.ID]; exists {
		depRoot = depNode
	} else {
		return errors.New("Dependency Id: " + dependency.ID + " not found")
	}

	var f reflect.Value

	if isPtr {
		f = reflect.Indirect(cp).FieldByName(dependency.Field)
	} else {
		f = cp.FieldByName(dependency.Field)
	}

	if f.IsValid() {

		if f.CanSet() {

			depcp, err := this.resolve(depRoot, nodeMap)
			if err != nil {
				return err
			}

			f.Set(reflect.ValueOf(depcp))

		} else {
			return errors.New("Field cannot be set")
		}

	} else {
		return errors.New("Invalid Field: " + dependency.Field + " from dep: " + depRoot.Type)
	}

	return nil
}
