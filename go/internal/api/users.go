package api

import (
	"path"
)

// ── Users ─────────────────────────────────────────────────────────────────────

type User struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	FullName  string `json:"fullName"`
}

type UsersResponse struct {
	Content []User `json:"content"`
}

func (c *Client) ListUsers() (*UsersResponse, error) {
	resp, err := c.get("/api/v1/users", nil)
	if err != nil {
		return nil, err
	}
	var result UsersResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ── User groups ───────────────────────────────────────────────────────────────

type UserGroup struct {
	UserGroupID string `json:"userGroupId"`
	Name        string `json:"name"`
	Users       []User `json:"users"`
}

type UserGroupsResponse struct {
	Content []UserGroup `json:"content"`
}

type CreateUserGroupRequest struct {
	Name          string   `json:"name"`
	MemberEmails  []string `json:"memberEmails,omitempty"`
}

type UpdateUserGroupRequest struct {
	Name               *string  `json:"name,omitempty"`
	AddMemberEmails    []string `json:"addMemberEmails,omitempty"`
	RemoveMemberEmails []string `json:"removeMemberEmails,omitempty"`
}

func (c *Client) ListUserGroups() (*UserGroupsResponse, error) {
	resp, err := c.get("/api/v1/userGroups", nil)
	if err != nil {
		return nil, err
	}
	var result UserGroupsResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateUserGroup(req CreateUserGroupRequest) (*UserGroup, error) {
	resp, err := c.post("/api/v1/userGroups", req)
	if err != nil {
		return nil, err
	}
	// API returns 201 + Location header, empty body — extract ID from Location
	loc := resp.Header.Get("Location")
	if decodeErr := decode(resp, &struct{}{}); decodeErr != nil {
		return nil, decodeErr
	}
	return &UserGroup{
		UserGroupID: path.Base(loc),
		Name:        req.Name,
	}, nil
}

func (c *Client) UpdateUserGroup(groupID string, req UpdateUserGroupRequest) (*UserGroup, error) {
	resp, err := c.post("/api/v1/userGroups/"+groupID, req)
	if err != nil {
		return nil, err
	}
	var result UserGroup
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteUserGroup(groupID string) error {
	resp, err := c.delete("/api/v1/userGroups/" + groupID)
	if err != nil {
		return err
	}
	var result struct{}
	return decode(resp, &result)
}
