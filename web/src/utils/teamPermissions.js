/**
 * 团队权限判断工具函数
 * 统一的三级权限体系：Owner (0) / Admin (1) / Member (2)
 */

/**
 * 角色常量定义
 */
export const ROLE = {
    OWNER: 0,
    ADMIN: 1,
    MEMBER: 2
};

/**
 * 判断是否为团队所有者
 * @param {number} role - 用户角色
 * @returns {boolean}
 */
export const isTeamOwner = (role) => {
    return role === ROLE.OWNER;
};

/**
 * 判断是否可以管理成员（Owner/Admin）
 * @param {number} role - 用户角色
 * @returns {boolean}
 */
export const canManageMembers = (role) => {
    return role === ROLE.OWNER || role === ROLE.ADMIN;
};

/**
 * 判断是否可以管理特定成员
 * Admin 可以管理 Member，但不能管理其他 Admin 或 Owner
 * @param {number} currentUserRole - 当前用户角色
 * @param {number} targetUserRole - 目标用户角色
 * @returns {boolean}
 */
export const canManageSpecificMember = (currentUserRole, targetUserRole) => {
    // Owner 可以管理所有人
    if (currentUserRole === ROLE.OWNER) return true;

    // Admin 只能管理 Member
    if (currentUserRole === ROLE.ADMIN) {
        return targetUserRole === ROLE.MEMBER;
    }

    // Member 不能管理任何人
    return false;
};

/**
 * 判断是否可以管理额度（仅 Owner）
 * @param {number} role - 用户角色
 * @returns {boolean}
 */
export const canManageQuota = (role) => {
    return role === ROLE.OWNER;
};

/**
 * 判断是否可以访问设置页（Owner/Admin）
 * @param {number} role - 用户角色
 * @returns {boolean}
 */
export const canAccessSettings = (role) => {
    return role === ROLE.OWNER || role === ROLE.ADMIN;
};

/**
 * 判断是否可以查看详细统计
 * @param {number} role - 用户角色
 * @param {Object} permissions - 团队权限配置
 * @returns {boolean}
 */
export const canViewDetailedStats = (role, permissions = {}) => {
    // Owner 和 Admin 始终可以查看
    if (role === ROLE.OWNER || role === ROLE.ADMIN) return true;

    // Member 根据团队配置决定
    return permissions.allowMemberViewStats !== false;
};

/**
 * 判断是否可以导出数据（Owner/Admin）
 * @param {number} role - 用户角色
 * @returns {boolean}
 */
export const canExportData = (role) => {
    return role === ROLE.OWNER || role === ROLE.ADMIN;
};

/**
 * 判断是否可以解散团队（仅 Owner）
 * @param {number} role - 用户角色
 * @returns {boolean}
 */
export const canDissolveTeam = (role) => {
    return role === ROLE.OWNER;
};

/**
 * 判断是否可以转让所有权（仅 Owner）
 * @param {number} role - 用户角色
 * @returns {boolean}
 */
export const canTransferOwnership = (role) => {
    return role === ROLE.OWNER;
};

/**
 * 判断是否可以修改团队状态（仅 Owner）
 * @param {number} role - 用户角色
 * @returns {boolean}
 */
export const canModifyTeamStatus = (role) => {
    return role === ROLE.OWNER;
};

/**
 * 判断是否可以查看所有成员活动（Owner/Admin）
 * @param {number} role - 用户角色
 * @returns {boolean}
 */
export const canViewAllActivities = (role) => {
    return role === ROLE.OWNER || role === ROLE.ADMIN;
};

/**
 * 判断成员是否可以查看其他成员的详细信息
 * @param {number} currentUserRole - 当前用户角色
 * @param {number} targetUserId - 目标用户ID
 * @param {number} currentUserId - 当前用户ID
 * @returns {boolean}
 */
export const canViewMemberDetails = (currentUserRole, targetUserId, currentUserId) => {
    // Owner 和 Admin 可以查看所有成员
    if (currentUserRole === ROLE.OWNER || currentUserRole === ROLE.ADMIN) {
        return true;
    }

    // Member 只能查看自己
    return targetUserId === currentUserId;
};

/**
 * 获取角色显示文本
 * @param {number} role - 角色值
 * @returns {string}
 */
export const getRoleText = (role) => {
    switch (role) {
        case ROLE.OWNER:
            return 'Owner';
        case ROLE.ADMIN:
            return 'Admin';
        case ROLE.MEMBER:
            return 'Member';
        default:
            return 'Unknown';
    }
};

/**
 * 获取角色 Badge 变体
 * @param {number} role - 角色值
 * @returns {string}
 */
export const getRoleBadgeVariant = (role) => {
    switch (role) {
        case ROLE.OWNER:
            return 'destructive'; // 红色
        case ROLE.ADMIN:
            return 'default'; // 蓝色
        case ROLE.MEMBER:
            return 'secondary'; // 灰色
        default:
            return 'outline';
    }
};

/**
 * 获取角色颜色（兼容旧代码）
 * @param {number} role - 角色值
 * @returns {string}
 * @deprecated 使用 getRoleBadgeVariant 替代
 */
export const getRoleColor = (role) => {
    if (role === ROLE.OWNER) return 'error';
    return role === ROLE.ADMIN ? 'primary' : 'default';
};
