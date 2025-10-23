/**
 * Team 权限判断工具函数
 */

/**
 * 判断用户是否有团队管理权限
 * @param {Object} team - 团队信息对象
 * @param {number} userId - 用户ID
 * @param {number} currentUserRole - 当前用户角色 (1=管理员, 2=普通成员)
 * @param {boolean} isOwner - 是否为团队所有者
 * @returns {boolean} 是否有管理权限
 */
export const hasManagePermission = (team, userId, currentUserRole, isOwner) => {
  // 如果是团队所有者，有管理权限
  if (isOwner) {
    return true;
  }
  
  // 如果是管理员角色，有管理权限
  if (currentUserRole === 1) {
    return true;
  }
  
  return false;
};

/**
 * 获取角色显示文本
 * @param {number} role - 角色值
 * @param {boolean} isOwner - 是否为所有者
 * @returns {string} 角色文本
 */
export const getRoleText = (role, isOwner) => {
  if (isOwner) {
    return 'Owner';
  }
  return role === 1 ? 'Admin' : 'Member';
};

/**
 * 获取角色颜色
 * @param {number} role - 角色值
 * @param {boolean} isOwner - 是否为所有者
 * @returns {string} 角色颜色
 */
export const getRoleColor = (role, isOwner) => {
  if (isOwner) {
    return 'error'; // 红色
  }
  return role === 1 ? 'primary' : 'default'; // 蓝色 : 灰色
};


