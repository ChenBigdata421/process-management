package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "process-management/infrastructure/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UserServiceClient 用户服务客户端
type UserServiceClient struct {
	conn       *grpc.ClientConn
	userClient pb.UserInfoServiceClient
	orgClient  pb.OrgInfoServiceClient
	tenantID   string
}

// NewUserServiceClient 创建用户服务客户端
func NewUserServiceClient(address string, tenantID string) (*UserServiceClient, error) {
	// 创建 gRPC 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %v", err)
	}

	log.Printf("[UserServiceClient] Connected to user service at %s", address)

	return &UserServiceClient{
		conn:       conn,
		userClient: pb.NewUserInfoServiceClient(conn),
		orgClient:  pb.NewOrgInfoServiceClient(conn),
		tenantID:   tenantID,
	}, nil
}

// Close 关闭连接
func (c *UserServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetUserByID 根据用户ID获取用户信息
func (c *UserServiceClient) GetUserByID(ctx context.Context, userID int32) (*pb.UserInfoReply, error) {
	req := &pb.GetUserByIdReq{
		TenantId: c.tenantID,
		UserId:   userID,
	}

	resp, err := c.userClient.GetUserById(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID %d: %v", userID, err)
	}

	return resp, nil
}

// GetUserByPoliceNo 根据警号获取用户信息
func (c *UserServiceClient) GetUserByPoliceNo(ctx context.Context, policeNo string) (*pb.UserInfoReply, error) {
	req := &pb.GetUserByPoliceNoReq{
		TenantId: c.tenantID,
		PoliceNo: policeNo,
	}

	resp, err := c.userClient.GetUserByPoliceNo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by police no %s: %v", policeNo, err)
	}

	return resp, nil
}

// GetOrgByID 根据组织ID获取组织信息
func (c *UserServiceClient) GetOrgByID(ctx context.Context, orgID int32) (*pb.OrgInfoReply, error) {
	req := &pb.GetOrgByIdReq{
		TenantId: c.tenantID,
		OrgId:    orgID,
	}

	resp, err := c.orgClient.GetOrgById(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get org by ID %d: %v", orgID, err)
	}

	return resp, nil
}

// GetOrgByCode 根据组织编码获取组织信息
func (c *UserServiceClient) GetOrgByCode(ctx context.Context, orgCode string) (*pb.OrgInfoReply, error) {
	req := &pb.GetOrgByCodeReq{
		TenantId: c.tenantID,
		OrgCode:  orgCode,
	}

	resp, err := c.orgClient.GetOrgByCode(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get org by code %s: %v", orgCode, err)
	}

	return resp, nil
}

// GetOrgByName 根据组织名称获取组织信息
func (c *UserServiceClient) GetOrgByName(ctx context.Context, orgName string) (*pb.OrgInfoReply, error) {
	req := &pb.GetOrgByNameReq{
		TenantId: c.tenantID,
		OrgName:  orgName,
	}

	resp, err := c.orgClient.GetOrgByName(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get org by name %s: %v", orgName, err)
	}

	return resp, nil
}

// GetOrgFullName 获取组织全名
func (c *UserServiceClient) GetOrgFullName(ctx context.Context, orgID int32) (string, error) {
	req := &pb.GetOrgFullNameReq{
		TenantId: c.tenantID,
		OrgId:    orgID,
	}

	resp, err := c.orgClient.GetOrgFullName(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get org full name for ID %d: %v", orgID, err)
	}

	return resp.FullName, nil
}

// UserInfo 用户信息（简化版）
type UserInfo struct {
	UserID    int32
	UserName  string
	PoliceNo  string
	Phone     string
	RoleID    int32
	RoleName  string
	OrgID     int32
	OrgName   string
	Email     string
	Status    string
}

// GetUserInfo 获取用户信息（简化版）
func (c *UserServiceClient) GetUserInfo(ctx context.Context, userID int32) (*UserInfo, error) {
	resp, err := c.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	info := &UserInfo{
		UserID:   resp.UserId,
		UserName: resp.UserName,
		PoliceNo: resp.PoliceNo,
		Phone:    resp.Phone,
		RoleID:   resp.RoleId,
		OrgID:    resp.OrgId,
		Email:    resp.Email,
		Status:   resp.Status,
	}

	if resp.Role != nil {
		info.RoleName = resp.Role.RoleName
	}

	if resp.Org != nil {
		info.OrgName = resp.Org.OrgName
	}

	return info, nil
}

// OrgInfo 组织信息（简化版）
type OrgInfo struct {
	OrgID      int32
	OrgName    string
	OrgCode    string
	OrgType    string
	ParentID   int32
	ParentName string
	Leader     string
	Phone      string
	Status     string
}

// GetOrgInfo 获取组织信息（简化版）
func (c *UserServiceClient) GetOrgInfo(ctx context.Context, orgID int32) (*OrgInfo, error) {
	resp, err := c.GetOrgByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	info := &OrgInfo{
		OrgID:    resp.OrgId,
		OrgName:  resp.OrgName,
		OrgCode:  resp.OrgCode,
		OrgType:  resp.OrgType,
		ParentID: resp.ParentId,
		Leader:   resp.Leader,
		Phone:    resp.Phone,
		Status:   resp.Status,
	}

	if resp.ParentOrg != nil {
		info.ParentName = resp.ParentOrg.OrgName
	}

	return info, nil
}

// ResolveUserID 解析用户标识符为用户ID
// 支持格式：
// - 数字：直接作为用户ID
// - 警号：通过警号查询用户
func (c *UserServiceClient) ResolveUserID(ctx context.Context, identifier string) (int32, error) {
	// 尝试解析为数字
	var userID int32
	_, err := fmt.Sscanf(identifier, "%d", &userID)
	if err == nil {
		return userID, nil
	}

	// 作为警号查询
	user, err := c.GetUserByPoliceNo(ctx, identifier)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve user identifier %s: %v", identifier, err)
	}

	return user.UserId, nil
}

// ResolveOrgID 解析组织标识符为组织ID
// 支持格式：
// - 数字：直接作为组织ID
// - 组织编码：通过编码查询组织
// - 组织名称：通过名称查询组织
func (c *UserServiceClient) ResolveOrgID(ctx context.Context, identifier string) (int32, error) {
	// 尝试解析为数字
	var orgID int32
	_, err := fmt.Sscanf(identifier, "%d", &orgID)
	if err == nil {
		return orgID, nil
	}

	// 尝试作为组织编码查询
	org, err := c.GetOrgByCode(ctx, identifier)
	if err == nil {
		return org.OrgId, nil
	}

	// 尝试作为组织名称查询
	org, err = c.GetOrgByName(ctx, identifier)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve org identifier %s: %v", identifier, err)
	}

	return org.OrgId, nil
}

