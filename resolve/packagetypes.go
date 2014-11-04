package resolve

import (
    "fmt"
    "github.com/andrewchambers/g/parse"
)



// convert a list of unordered type decls into GNamedTypes. Self referencing types
// are allowed indirectly through pointers.
// Will detect:
//   redefinition of a type.
//   use of a non type symbol in a type position. 

func getTopLevelNamedTypes(decls []*parse.TypeDecl) ([]*GNamedType,error) {
    
    ret := make([]*GNamedType,0,len(decls))
    tyLookup := make(map[string] *GNamedType)
    tdLookup := make(map[string] *parse.TypeDecl)
    
    // For each type decl create a GNamedType.
    
    for _,td := range decls {
        _,ok := tyLookup[td.Name]
        if ok {
            return ret,fmt.Errorf("redefinition of type %s at %s",td.Name,td.GetSpan().Start)
        }
        t := &GNamedType{}
        t.Name = td.Name
        ret = append(ret,t)
        tyLookup[td.Name] = t
        tdLookup[td.Name] = td
    }
    
    lookup := func (i *parse.Ident) (*GNamedType,error) {
        name := i.Val
        ty,ok := tyLookup[name]
        if !ok {
            return nil,fmt.Errorf("undefined type %s at %s:%s",name,i.GetSpan().Path,i.GetSpan().Start)
        }
        return ty,nil
    }
    
    // For each Type decl, recursively create the types.
    
    for _,td := range decls {
        t, err := astNodeToGType(lookup, td.Type)
        tyLookup[td.Name].Type = t
        if err != nil {
            return ret, err
        }
    }

    
    // For each GType ensure it does not contain itself in a non reference form.
    
    for _,ty := range ret {
        if containsInvalidTypeRecursion(ty,ty.Type) {
            td := tdLookup[ty.Name]
            return ret,fmt.Errorf("self recursive type %s:%s at ",td.GetSpan().Path,td.GetSpan().Start)
        }
    }
    
    
    return ret, nil
}


func containsInvalidTypeRecursion(named *GNamedType, t GType) bool {
    switch t := t.(type) {
        case *GPointer:
            return false
        case *GArray:
            return containsInvalidTypeRecursion(named,t.SubType)
        case *GStruct:
            for _,ty := range t.Types {
                if containsInvalidTypeRecursion(named,ty) {
                    return true
                }
            }
            return false
        case *GNamedType:
            return t == named
    }
    panic(t)
}



