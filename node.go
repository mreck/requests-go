package requests

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

func (p Node) TagName() (string, bool) {
	if p.root.Type == html.ElementNode {
		return p.root.Data, true
	}
	return "", false
}

func (p Node) Attr(key string) (string, bool) {
	if p.root.Type == html.ElementNode {
		for _, attr := range p.root.Attr {
			if attr.Key == key {
				return attr.Val, true
			}
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

func (p Node) GetElementsByID(id string) (Node, error) {
	var result Node

	p.WalkNodes(func(node Node) (more bool, err error) {
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

	p.WalkNodes(func(node Node) (more bool, err error) {
		if list, ok := node.ClassList(); ok && slices.Contains(list, name) {
			result = append(result, node)
		}
		return true, nil
	})

	return result, nil
}

func (p Node) GetElementsByTagName(name string) ([]Node, error) {
	var result []Node

	p.WalkNodes(func(node Node) (more bool, err error) {
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
