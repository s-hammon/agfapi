package agfa

import (
	"strings"

	"github.com/s-hammon/p"
)

type List struct {
	Code         Code
	Entry        []ListEntry
	Extension    []Extension
	Id           string
	Identifier   []ResourceIdentifier
	Meta         ResourceMeta
	Mode         string
	ResourceType string
	Status       string
	Title        string
}

type Bundle struct {
	ResourceType string
	Type         string
	Total        int
	Link         []BundleLink
	Entry        []BundleEntry
}

func (b Bundle) String() string {
	return p.Format("resource=%s, type=%s, total=%d\n", b.ResourceType, b.Type, b.Total)
}

func (b Bundle) Entries() []ListEntry {
	for _, entry := range b.Entry {
		if entry.Resource.ResourceType == "List" {
			return entry.Resource.Entry
		}
	}

	return nil
}

type BundleLink struct {
	Relation string
	Url      string
}

type BundleEntry struct {
	FullUrl  string
	Resource BundleEntryResource
}

type BundleEntryResource struct {
	ResourceType string
	Id           string
	Meta         ResourceMeta
	Extension    []UrlExtension
	Identifier   []ResourceIdentifier
	Status       string
	Mode         string
	Title        string
	Code         Code
	Entry        []ListEntry
}

type ResourceMeta struct {
	Profile []string
}

type UrlExtension struct {
	Url       string
	ValueCode string
	Extension []Extension
}

type Extension struct {
	Url          string
	ValueInteger int
	ValueBoolean bool
}

type ResourceIdentifier struct {
	Type   UrlExtensionIdentifierType
	System string
	Value  string
}

type UrlExtensionIdentifierType struct {
	Coding []UrlExtensionIdentifierTypeCoding
}

type UrlExtensionIdentifierTypeCoding struct {
	System  string
	Code    string
	Display string
}

type ListResource struct {
	ResourceType string
	Id           string
	Title        string
	Status       string
	Mode         string
	Code         Code
	Entry        []ListEntry
}

func (list ListResource) String() string {
	sb := strings.Builder{}
	sb.WriteString(p.Format("resourceType=%s\n", list.ResourceType))
	sb.WriteString(p.Format("id=%s\n", list.Id))
	sb.WriteString(p.Format("title=%s\n", list.Title))
	sb.WriteString(p.Format("status=%s\n", list.Status))
	sb.WriteString(p.Format("mode=%s\n", list.Mode))
	sb.WriteString(list.Code.String())

	return sb.String()
}

type Coding struct {
	System  string
	Code    string
	Display string
}

func (c Coding) String() string {
	return p.Format("system=%s, code=%s, display=%s", c.System, c.Code, c.Display)
}

type Code struct {
	Coding []Coding
	Text   string
}

func (cc Code) String() string {
	sb := strings.Builder{}
	sb.WriteString("coding:\n")
	for _, c := range cc.Coding {
		sb.WriteString(p.Format("\t%s\n", c))
	}

	sb.WriteString(p.Format("text=%s\n", cc.Text))
	return sb.String()
}

type ListEntry struct {
	Item ListEntryItem
}

func (le ListEntry) String() string {
	return p.Format("item.reference=%s", le.Item.Reference)
}

type ListEntryItem struct {
	Reference string
}

func (le ListEntryItem) IsTask() bool {
	return strings.HasPrefix(le.Reference, "Task/")
}

func (le ListEntryItem) ExtractTaskId() string {
	return le.Reference[strings.Index(le.Reference, "/")+1:]
}

type Task struct {
	ResourceType string
	Id           string
	Identifier   []ResourceIdentifier
	Status       string
	Intent       string
	Priority     string
	Code         Code
	For          Reference
	AuthoredOn   string
	LastModified string
	Input        []TaskInput
}

func (t Task) ServiceRequestId() string {
	if len(t.Input) == 0 {
		return ""
	}

	ref := t.Input[0].ValueReference.Reference
	return ref[strings.Index(ref, "/")+1:]
}

type Reference struct {
	Reference string
	Display   string
}

type TaskInput struct {
	ValueReference Reference
}

type ServiceRequest struct {
	ResourceType       string
	Id                 string
	Identifier         []ResourceIdentifier
	Status             string
	Intent             string
	Priority           string
	Code               Code
	Subject            Reference
	Encounter          Reference
	OccurrenceDateTime string
	Performer          []Reference
}
