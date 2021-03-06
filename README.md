### A simple Go package to find the nested JSON Data

#### Installation
```
go get github.com/krystalics/jpath@v0.0.3
```

#### Usage
```
import "github.com/krystalics/jpath"
```

Let's see a quick example:

```go
package main
import  "github.com/krystalics/jpath"

func main() {
	const json = `{"name":{"first":"lin","last":"jia"},"age":61}`
	jPath, _ := jpath.New(json)
	name,err := jPath.Find("name.first")
	if err != nil {
		return
	}
	//...
}
```

when you want to use jpath in concurrency situation
```go
package main
import  "github.com/krystalics/jpath"

func main() {
	const json = `{"name":{"first":"lin","last":"jia"},"age":61}`
	jPath, _ := jpath.NewConcurrencySafe(json)
	for i := 0; i < 100; i++ {
		go func() {
			name,err := jPath.Find("name.first")
			if err != nil {
				return
			}
			//...
		}()
	}
}
```

#### LICENSE
The jpath is an open-source software licensed under the MIT License.