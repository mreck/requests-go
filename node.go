package requestsgo

import (
	"io"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

type Node struct {
	root *html.Node
}

func ParseHTML(r io.Reader) (Node, error) {
	root, err := html.Parse(r)
	if err != nil {
		return Node{}, err
	}
	return Node{root: root}, nil
}

func (p Node) Node() *html.Node {
	return p.root
}

func (p Node) IsElement() bool {
	return p.root.Type == html.ElementNode
}

func (p Node) TagName() (string, bool) {
	if !p.IsElement() {
		return "", false
	}
	return p.root.Data, true
}

func (p Node) Attr(key string) (string, bool) {
	if !p.IsElement() {
		return "", false
	}
	for _, attr := range p.root.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

func (p Node) ID() (string, bool) {
	return p.Attr("id")
}

func (p Node) ClassList() ([]string, bool) {
	classStr, ok := p.Attr("class")
	if !ok {
		return nil, false
	}
	result := strings.Split(classStr, " ")
	for i, class := range result {
		result[i] = strings.TrimSpace(class)
	}
	return result, true
}

func (p Node) DataSet() (map[string]string, bool) {
	if !p.IsElement() {
		return nil, false
	}
	result := map[string]string{}
	for _, attr := range p.root.Attr {
		if strings.HasPrefix(attr.Key, "data-") {
			result[strings.TrimPrefix(attr.Key, "data-")] = attr.Val
		}
	}
	return result, true
}

func (p Node) WalkNodes(fn func(node Node) (more bool, err error)) error {
	_, err := p.walkNodes(p.root, fn)
	return err
}

func (p Node) walkNodes(node *html.Node, fn func(node Node) (more bool, err error)) (bool, error) {
	more, err := fn(Node{node})
	if err != nil {
		return false, err
	}
	if !more {
		return false, nil
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		more, err := p.walkNodes(child, fn)
		if err != nil {
			return false, err
		}
		if !more {
			return false, nil
		}
	}

	return true, nil
}

func (p Node) WalkElements(fn func(node Node) (more bool, err error)) error {
	_, err := p.walkNodes(p.root, func(node Node) (more bool, err error) {
		if !node.IsElement() {
			return true, nil
		}
		return fn(node)
	})
	return err
}

func (p Node) GetElementByID(id string) (Node, error) {
	var result Node

	p.WalkElements(func(node Node) (more bool, err error) {
		if id, ok := node.ID(); ok && id == "id" {
			result = node
			return false, nil
		}
		return true, nil
	})

	return result, nil
}

func (p Node) GetElementsByClassName(name string) ([]Node, error) {
	var result []Node

	p.WalkElements(func(node Node) (more bool, err error) {
		if list, ok := node.ClassList(); ok && slices.Contains(list, name) {
			result = append(result, node)
		}
		return true, nil
	})

	return result, nil
}

func (p Node) GetElementsByTagName(name string) ([]Node, error) {
	var result []Node

	p.WalkElements(func(node Node) (more bool, err error) {
		if tagName, ok := node.TagName(); ok && tagName == name {
			result = append(result, node)
		}
		return true, nil
	})

	return result, nil
}

func (p Node) GetLinks() ([]string, error) {
	nodes, err := p.GetElementsByTagName("a")
	if err != nil {
		return nil, err
	}

	var result []string

	for _, node := range nodes {
		if href, ok := node.Attr("href"); ok {
			result = append(result, href)
		}
	}

	return result, nil
}
