package mempage

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joy717/mempage/reflections"
)

type Operation string

const (
	Like    Operation = "like" //模糊查詢
	Eq      Operation = "eq"   //精確匹配
	Ne      Operation = "ne"
	In      Operation = "in"
	NotIn   Operation = "not in"
	IsNull  Operation = "is null"
	NotNull Operation = "not null"
)

var defaultPageSize = 10

type Page struct {
	Page         int          `json:"page" example:"1"`      //当前页
	PageSize     int          `json:"pageSize" example:"10"` //每页数量
	Filters      []PageFilter `json:"filters"`               //过滤
	Sorts        []PageSort   `json:"sorts"`                 //排序
	Result       interface{}  `json:"result"`                //分页数据
	TotalCount   int          `json:"totalCount"`            //总条数
	mGenericData []interface{}
}

type PageFilter struct {
	Key    string    `json:"key" example:"name"` //过滤项(json tag)，深度过滤用.隔开：cpu.total
	Op     Operation `json:"op" example:"like"`  //操作類型 like、eq、is null、not null、>、>=、<、<=、?=、!?=
	Values []string  `json:"values" example:""`  //过滤值
}

type PageSort struct {
	Key       string `json:"key" example:"name"`      //排序项，深度排序用.隔开：cpu.total
	Ascending bool   `json:"asceding" example:"true"` //是否升序
}

func (p *Page) Len() int { return len(p.mGenericData) }

func (p *Page) Swap(i, j int) {
	p.mGenericData[i], p.mGenericData[j] = p.mGenericData[j], p.mGenericData[i]
}

func (p *Page) Less(i, j int) bool {
	for _, s := range p.Sorts {
		sortBy := s.Key
		pa := p.getProperty(p.mGenericData[i], sortBy)
		pb := p.getProperty(p.mGenericData[j], sortBy)
		if pa == nil || pb == nil {
			break
		}
		cmp := p.compare(pa, pb)
		if cmp == 0 { // values are the same. Just continue to next sortBy
			continue
		} else { // values different
			return (cmp == -1 && s.Ascending) || (cmp == 1 && !s.Ascending)
		}
	}
	return false
}

func (p *Page) sort() *Page {
	sort.Stable(p)
	return p
}

func (p *Page) paginate() *Page {
	if p.PageSize == 0 {
		p.PageSize = defaultPageSize
	}

	if p.Page < 1 || p.PageSize < 1 {
		return p
	}

	startIndex := p.PageSize * (p.Page - 1)
	endIndex := p.PageSize * p.Page

	if endIndex > p.TotalCount {
		endIndex = p.TotalCount
	}

	if startIndex > p.TotalCount {
		startIndex = p.TotalCount
	}

	p.Result = p.mGenericData[startIndex:endIndex]
	return p
}

func (p *Page) GetmGenericData() []interface{} {
	return p.mGenericData
}

func (p *Page) filter() *Page {
	count := len(p.mGenericData)
	shouldFilter := len(p.Filters) > 0
	filteredList := make([]interface{}, 0)

	for i := 0; i < count; i++ {
		obj := p.mGenericData[i]
		if !shouldFilter || p.match(obj) {
			filteredList = append(filteredList, obj)
		}
	}

	p.mGenericData = filteredList
	p.TotalCount = p.Len()
	return p
}

func (p *Page) compare(objA, objB interface{}) int {
	typeA := reflect.TypeOf(objA)
	typeB := reflect.TypeOf(objB)
	if typeA.String() != typeB.String() {
		return 0
	}

	if typeA.Kind() == reflect.String {
		return strings.Compare(strings.ToUpper(fmt.Sprintf("%v", objA)), strings.ToUpper(fmt.Sprintf("%v", objB)))
	} else if typeA.Kind() == reflect.Int32 {
		return p.int32Compare(objA.(int32), objB.(int32))
	} else if typeA.Kind() == reflect.Int {
		return p.intCompare(objA.(int), objB.(int))
	} else if typeA.Kind() == reflect.Int64 {
		return p.int64Compare(objA.(int64), objB.(int64))
	} else if typeA.Kind() == reflect.Float32 {
		return p.float32Compare(objA.(float32), objB.(float32))
	} else if typeA.Kind() == reflect.Float64 {
		return p.float64Compare(objA.(float64), objB.(float64))
	} else if typeA == reflect.TypeOf(time.Time{}) {
		timeA := objA.(time.Time)
		timeB := objB.(time.Time)

		return p.int64Compare(timeA.Unix(), timeB.Unix())
		//} else if typeA == reflect.TypeOf(&timestamp.Timestamp{}) { //proto buf timestamp compare
		//	timestampA := objA.(*timestamp.Timestamp)
		//	timestampB := objp.(*timestamp.Timestamp)
		//	return p.timestampCompare(timestampA, timestampB)
	}
	return 0
}

func (p *Page) intCompare(x, y int) int {
	if x > y {
		return 1
	} else if x == y {
		return 0
	}
	return -1
}

func (p *Page) int32Compare(x, y int32) int {
	if x > y {
		return 1
	} else if x == y {
		return 0
	}
	return -1
}

func (p *Page) int64Compare(x, y int64) int {
	if x > y {
		return 1
	} else if x == y {
		return 0
	}
	return -1
}

func (p *Page) float64Compare(x, y float64) int {
	if x > y {
		return 1
	} else if x == y {
		return 0
	}
	return -1
}
func (p *Page) float32Compare(x, y float32) int {
	if x > y {
		return 1
	} else if x == y {
		return 0
	}
	return -1
}

//func (p *Page) timestampCompare(a, b *timestamp.Timestamp) int {
//	if a == nil || b == nil {
//		return -1
//	}
//
//	aTime, _ := ptypes.Timestamp(a)
//	bTime, _ := ptypes.Timestamp(b)
//	if aTime.After(bTime) {
//		return 1
//	} else if aTime.Equal(bTime) {
//		return 0
//	}
//	return -1
//}

func (p *Page) match(obj interface{}) bool {
	for _, f := range p.Filters {
		val := p.getProperty(obj, f.Key)
		if val == nil {
			return false
		}
		matched := false
		if bo, ok := val.(bool); ok {
			for _, v := range f.Values {
				if tmp, _ := strconv.ParseBool(v); bo == tmp {
					matched = true
				}
			}
		} else if bo, ok := val.(int64); ok {
			for _, v := range f.Values {
				if tmp, _ := strconv.ParseInt(v, 10, 64); bo == tmp {
					matched = true
				}
			}
		} else {
			str := ""
			if reflect.TypeOf(val).Kind() == reflect.String {
				str = reflect.ValueOf(val).String()
			}
			/*str, ok := val.(string)
			if !ok {
				return false
			}*/
			switch f.Op {
			case IsNull:
				return str == ""
			case NotNull:
				return str != ""
			case Eq:
				for _, v := range f.Values {
					if str == v {
						matched = true
						break
					}
				}
			case In:
				for _, v := range f.Values {
					vs := strings.Split(v, ",")
					for _, v1 := range vs {
						if strings.Contains(str, v1) {
							matched = true
							break
						}
					}
				}
			case NotIn:
				flag := false
				for _, v := range f.Values {
					vs := strings.Split(v, ",")
					for _, v1 := range vs {
						if str == v1 {
							flag = true
							break
						}
					}
				}
				if flag {
					matched = false
				} else {
					matched = true
				}
			default:
				for _, v := range f.Values {
					if strings.Contains(str, v) {
						matched = true
						break
					}
				}
			}

		}
		if !matched {
			return false
		}
	}

	return true
}

func (p *Page) getProperty(obj interface{}, name string) interface{} {
	o := obj
	for _, n := range strings.Split(name, ".") {
		o = p.doGetProperty(o, n)
		if o == nil {
			return nil
		}
	}
	return o
}

func (p *Page) doGetProperty(obj interface{}, name string) interface{} {
	k, err := reflections.GetFieldNameByJsonTag(obj, name)
	if err != nil {
		fmt.Printf("obj < %+v >does not have json tag:%v,error is :%v\n", obj, name, err)
		return nil
	}

	v, err := reflections.GetField(obj, k)
	if err != nil {
		fmt.Printf("obj < %+v >does not have property:%v,error is :%v\n", obj, k, err)
		return nil
	}

	return v
}

func (p *Page) FillResultAll(result interface{}) *Page {
	if reflect.TypeOf(result).Kind() == reflect.Slice {
		values := reflect.ValueOf(result)
		count := values.Len()

		filteredList := make([]interface{}, 0)

		for i := 0; i < count; i++ {
			filteredList = append(filteredList, values.Index(i).Interface())
		}

		p.mGenericData = filteredList
	} else {
		p.mGenericData = []interface{}{result}
	}

	return p.filter().sort().paginate()
}
