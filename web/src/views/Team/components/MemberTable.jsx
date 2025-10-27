import { useState, useMemo } from 'react';
import PropTypes from 'prop-types';
import { Search } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Card, CardContent } from '@/components/ui/card';
import { renderQuota } from 'utils/common';
import { getRoleText, getRoleBadgeVariant, ROLE } from 'utils/teamPermissions';
import MemberActionMenu from './MemberActionMenu';

const MemberTable = ({ members, currentUserRole, onSetQuota, onChangeRole, onRemove }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState('all');

  // 过滤成员
  const filteredMembers = useMemo(() => {
    return members.filter((member) => {
      // 搜索过滤
      if (searchTerm) {
        const username = member.user?.username?.toLowerCase() || '';
        const email = member.user?.email?.toLowerCase() || '';
        const search = searchTerm.toLowerCase();
        if (!username.includes(search) && !email.includes(search)) {
          return false;
        }
      }

      // 角色过滤
      if (roleFilter !== 'all') {
        if (roleFilter === 'owner' && member.role !== ROLE.OWNER) return false;
        if (roleFilter === 'admin' && member.role !== ROLE.ADMIN) return false;
        if (roleFilter === 'member' && member.role !== ROLE.MEMBER) return false;
      }

      return true;
    });
  }, [members, searchTerm, roleFilter]);

  // 格式化时间
  const formatTime = (timestamp) => {
    if (!timestamp) return '-';
    return new Date(timestamp * 1000).toLocaleDateString('zh-CN');
  };

  // 计算使用率
  const calculateUsageRate = (used, max, unlimited) => {
    if (unlimited || max === 0) return 0;
    return (used / max) * 100;
  };

  // 获取用户名首字母
  const getUserInitial = (username) => {
    return username?.charAt(0).toUpperCase() || 'U';
  };

  if (members.length === 0) {
    return (
      <Card className="border-dashed">
        <CardContent className="flex flex-col items-center justify-center py-12">
          <p className="text-muted-foreground">暂无团队成员</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* 搜索和筛选 */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
          <Input
            placeholder="搜索成员（用户名或邮箱）..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
        <Select value={roleFilter} onValueChange={setRoleFilter}>
          <SelectTrigger className="w-full sm:w-32">
            <SelectValue placeholder="角色" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部角色</SelectItem>
            <SelectItem value="owner">所有者</SelectItem>
            <SelectItem value="admin">管理员</SelectItem>
            <SelectItem value="member">普通成员</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* 成员表格 */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>成员</TableHead>
              <TableHead>角色</TableHead>
              <TableHead>额度限制</TableHead>
              <TableHead>额度使用</TableHead>
              <TableHead>加入时间</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredMembers.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">
                  没有找到符合条件的成员
                </TableCell>
              </TableRow>
            ) : (
              filteredMembers.map((member) => {
                const usageRate = calculateUsageRate(
                  member.used_quota || 0,
                  member.max_quota || 0,
                  member.unlimited_quota || member.max_quota === 0
                );

                return (
                  <TableRow key={member.id || member.user_id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <Avatar className="h-10 w-10">
                          <AvatarFallback>{getUserInitial(member.user?.username)}</AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="font-medium">{member.user?.username || '未知用户'}</div>
                          <div className="text-sm text-muted-foreground">{member.user?.email || '无邮箱'}</div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={getRoleBadgeVariant(member.role)}>{getRoleText(member.role)}</Badge>
                    </TableCell>
                    <TableCell>
                      {member.unlimited_quota || member.max_quota === 0 ? (
                        <span className="text-muted-foreground">无限制</span>
                      ) : (
                        renderQuota(member.max_quota, 2)
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="space-y-1 min-w-[150px]">
                        <div className="flex items-center justify-between text-sm">
                          <span>{renderQuota(member.used_quota || 0, 2)}</span>
                          {!member.unlimited_quota && member.max_quota > 0 && (
                            <span className="text-muted-foreground">{usageRate.toFixed(0)}%</span>
                          )}
                        </div>
                        {!member.unlimited_quota && member.max_quota > 0 && <Progress value={usageRate} className="h-1" />}
                      </div>
                    </TableCell>
                    <TableCell>{formatTime(member.joined_time)}</TableCell>
                    <TableCell className="text-right">
                      <MemberActionMenu
                        member={member}
                        currentUserRole={currentUserRole}
                        onSetQuota={onSetQuota}
                        onChangeRole={onChangeRole}
                        onRemove={onRemove}
                      />
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>

      {/* 统计信息 */}
      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>共 {filteredMembers.length} 个成员</span>
        {searchTerm || roleFilter !== 'all' ? <span>（已过滤 {members.length - filteredMembers.length} 个）</span> : null}
      </div>
    </div>
  );
};

MemberTable.propTypes = {
  members: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.number,
      user_id: PropTypes.number,
      role: PropTypes.number,
      max_quota: PropTypes.number,
      used_quota: PropTypes.number,
      unlimited_quota: PropTypes.bool,
      joined_time: PropTypes.number,
      user: PropTypes.shape({
        username: PropTypes.string,
        email: PropTypes.string
      })
    })
  ).isRequired,
  currentUserRole: PropTypes.number.isRequired,
  onSetQuota: PropTypes.func,
  onChangeRole: PropTypes.func,
  onRemove: PropTypes.func
};

export default MemberTable;
