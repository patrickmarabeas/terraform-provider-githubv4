package github

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/shurcooL/githubv4"
)

func dataSourceGithubOrganizationMembers() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			// Computed
			ORGANIZATION_MEMBERS: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						USER_LOGIN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						USER_IS_SITE_ADMIN: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						USER_NAME: {
							Type:     schema.TypeString,
							Computed: true,
						},
						USER_ROLE: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},

		Read: resourceGithubAppInitOrganizationMembersRead,
	}
}

func resourceGithubAppInitOrganizationMembersRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Organization struct {
			MembersWithRole struct {
				Edges []struct {
					Node User
					Role githubv4.OrganizationMemberRole
				}
				PageInfo PageInfo
			} `graphql:"membersWithRole(first: $first, after: $cursor)"`
			ID githubv4.ID
		} `graphql:"organization(login: $login)"`
	}
	variables := map[string]interface{}{
		"login":  githubv4.String(meta.(*Organization).Name),
		"first":  githubv4.Int(100),
		"cursor": (*githubv4.String)(nil),
	}

	var allEdges []struct {
		Node User
		Role githubv4.OrganizationMemberRole
	}
	ctx := context.Background()
	client := meta.(*Organization).Client
	for {
		err := client.Query(ctx, &query, variables)
		if err != nil {
			return err
		}
		allEdges = append(allEdges, query.Organization.MembersWithRole.Edges...)

		if !query.Organization.MembersWithRole.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.MembersWithRole.PageInfo.EndCursor)
	}

	var allUsers []map[string]interface{}
	for _, u := range allEdges {
		user := make(map[string]interface{})
		user[USER_IS_SITE_ADMIN] = bool(u.Node.IsSiteAdmin)
		user[USER_LOGIN] = string(u.Node.Login)
		user[USER_NAME] = string(u.Node.Name)
		user[USER_ROLE] = string(u.Role)
		allUsers = append(allUsers, user)
	}

	err := d.Set(ORGANIZATION_MEMBERS, allUsers)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s", query.Organization.ID))

	return nil
}
