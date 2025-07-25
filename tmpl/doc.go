package tmpl

import (
	"codegen/lang"
	"fmt"
	"github.com/samber/lo"
	"strings"
)

type Api struct {
	Name        string
	Description string
	Paths       []*Path
}

type Path struct {
	Tag          string
	Name         string
	Description  string
	Summary      string
	OriginalPath string
	Path         string
	Method       string
	Parameters   Parameters
	Queries      Parameters
	Request      *NamedType
	Response     *NamedType
}

type Output struct {
	Header    []string          `json:"header,omitempty"`    //文件头信息
	File      string            `json:"file,omitempty"`      //文件地址
	Ignore    bool              `json:"ignore,omitempty"`    //忽略生成
	Template  string            `json:"template,omitempty"`  //模版文件
	Variables map[string]string `json:"variables,omitempty"` //变量集
}

type Parameter struct {
	Name        string
	Alias       string
	Required    bool
	In          string
	Type        *NamedType
	Format      string
	Description string
	Default     string
}

type NamedTypeKind uint

const (
	ImmutableType NamedTypeKind = 1 << iota

	FoundationType

	MapType

	ArrayType

	ReferenceType

	GenericType

	RenameType

	VoidType
)

type NamedType struct {
	Kind       NamedTypeKind
	Expression string
}

var VoidNamedType = &NamedType{Kind: VoidType, Expression: ""}

type Parameters = []*Parameter

func (nt *NamedType) RenameExpression(scope string, name string, types map[string]string, convert lang.TypeConvert) {

	matches := []string{
		name,                              //全匹配
		fmt.Sprintf("%s:%s", scope, name), //配合属性对象，或者方法名
	}

	for k, v := range types {

		//~Id xxxId
		//Id~ Idxxx
		//~Id orderId
		matchPrefix := strings.HasPrefix(k, "~") && strings.HasSuffix(name, strings.TrimPrefix(k, "~"))
		matchSuffix := strings.HasSuffix(k, "~") && strings.HasPrefix(name, strings.TrimSuffix(k, "~"))

		//前后缀匹配
		if matchPrefix || matchSuffix || lo.Contains(matches, k) {

			nt.Kind = RenameType
			nt.Expression = v

			return
		}
	}
}

func (nt *NamedType) GenerateExpression(format string, convert lang.TypeConvert) {
	expression := nt.Expression

	if nt.Kind&ImmutableType != 0 {
		return
	}
	nt.Expression = nt.Kind.Parse(expression, format, convert)
}

func (nk NamedTypeKind) Parse(expression string, format string, convert lang.TypeConvert) string {

	if nk&FoundationType != 0 || nk&VoidType != 0 {
		expression = convert.Foundation(expression, format)
	}
	if nk&ReferenceType != 0 {
		expression = convert.Reference(expression)
	}
	if nk&ArrayType != 0 {
		expression = convert.Array(expression)
	}
	if nk&MapType != 0 {
		expression = convert.Map(expression)
	}

	return expression
}

type Ref struct {
	Name        string
	Alias       string
	Type        *NamedType
	Properties  Properties
	Description string
	Summary     string
	Ignore      bool
}

func (r *Ref) ReferenceLevel() int {

	return lo.Reduce(r.Properties, func(agg int, item *Property, index int) int {
		if item.Type.Kind&ReferenceType != 0 {
			return agg + 1
		}
		return agg
	}, 0)

}

type Properties []*Property

func (p Properties) Find(name string) (*Property, bool) {
	return lo.Find(p, func(item *Property) bool {
		return item.Name == name
	})
}

type Generic struct {
	Expression string
	Property   string
}

type Property struct {
	Name        string
	Alias       string
	Description string
	Summary     string
	Type        *NamedType
	Format      string
	Enums       []string
}
