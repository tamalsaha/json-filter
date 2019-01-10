package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func Filter(obj map[string]interface{}, filter map[string]interface{}) (map[string]interface{}, error) {
	return applyFilter(obj, filter, "")
}

func applyFilter(obj map[string]interface{}, filter map[string]interface{}, path string) (map[string]interface{}, error) {
	if obj == nil {
		return nil, nil
	}

	out := make(map[string]interface{}, len(obj))
	for k, subFilter := range filter {
		v, ok := obj[k]
		if !ok {
			continue // ignore missing key or throw error
		}
		sf, ok := subFilter.(map[string]interface{})
		if !ok {
			out[k] = v // just keep it as is
		} else {
			// apply sub filter
			// if v is an map, apply sub filter directly
			// else if v is an array of objects, apply to sub filter to individual elements
			// else, throw an error (filter is trying to apply to non objects)

			switch u := v.(type) {
			case map[string]interface{}:
				subOut, err := applyFilter(u, sf, path+k+".")
				if err != nil {
					return nil, err
				}
				out[k] = subOut
			case []interface{}:
				for i := range u {
					entry, ok := u[i].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("can't apply filter %v on %s%s[%d]: %v", sf, path, k, i, u[i]) // report the path to v
					}
					subOut, err := applyFilter(entry, sf, fmt.Sprintf("%s%s[%d].", path, k, i))
					if err != nil {
						return nil, err
					}
					u[i] = subOut
				}
				out[k] = u
			default:
				return nil, fmt.Errorf("can't apply filter %v on %s%s: %v", sf, path, k, v)
			}
		}
	}
	return out, nil
}

func main() {
	strObj := `
{
  "apiVersion": "extensions/v1beta1",
  "kind": "DaemonSet",
  "metadata": {
    "name": "busy-dm",
    "namespace": "default",
    "labels": {
      "app": "busy-dm"
    }
  },
  "spec": {
    "template": {
      "metadata": {
        "labels": {
          "name": "busy-dm"
        }
      },
      "spec": {
        "nodeSelector": {
          "kubernetes.io/hostname": "ip-172-20-53-35.ec2.internal"
        },
        "containers": [
          {
            "image": "busybox",
            "command": [
              "sleep",
              "3600"
            ],
            "imagePullPolicy": "IfNotPresent",
            "name": "busybox"
          },
          {
            "image": "nginx",
            "command": [
              "sleep",
              "3600"
            ],
            "imagePullPolicy": "IfNotPresent",
            "name": "nginx"
          }
        ]
      }
    }
  }
}`
	strFilter := `
{
  "apiVersion": null,
  "kind": null,
  "metadata": {
	"name": null,
	"namespace": null,
    "labels": {
      "app2": null
    }
  },
  "spec": {
    "template": {
      "spec": {
        "containers": {
          "name": null,
          "image": null
        }
      }
    }
  }
}`
	var obj map[string]interface{}
	err := json.Unmarshal([]byte(strObj), &obj)
	if err != nil {
		log.Fatalln(err)
	}

	var filter map[string]interface{}
	err = json.Unmarshal([]byte(strFilter), &filter)
	if err != nil {
		log.Fatalln(err)
	}

	out, err := Filter(obj, filter)
	if err != nil {
		log.Fatalln(err)
	}
	outBytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(outBytes))
	/*
	   {
	     "apiVersion": "extensions/v1beta1",
	     "kind": "DaemonSet",
	     "metadata": {
	       "labels": {},
	       "name": "busy-dm",
	       "namespace": "default"
	     },
	     "spec": {
	       "template": {
	         "spec": {
	           "containers": [
	             {
	               "image": "busybox",
	               "name": "busybox"
	             },
	             {
	               "image": "nginx",
	               "name": "nginx"
	             }
	           ]
	         }
	       }
	     }
	   }
	*/
}
