import { useState, useEffect, useCallback, useRef } from 'react';
import PropTypes from 'prop-types';
import { UserPlus, Search } from 'lucide-react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { API } from 'utils/api';
import { showError, showSuccess } from 'utils/common';
import { cn } from '@/lib/utils';

const InviteMemberDialog = ({ open, onOpenChange, teamId, onSuccess }) => {
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searchUsers, setSearchUsers] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [searching, setSearching] = useState(false);
  const [inviting, setInviting] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const debounceTimerRef = useRef(null);

  // 搜索用户
  const performSearch = useCallback(async (keyword) => {
    if (!keyword.trim()) {
      setSearchUsers([]);
      setHasSearched(false);
      return;
    }

    setSearching(true);
    setHasSearched(true);
    try {
      const response = await API.get(`/api/team/search_users?keyword=${encodeURIComponent(keyword)}`);
      if (response.data.success) {
        const usersData = response.data.data;
        let usersList = [];
        if (Array.isArray(usersData)) {
          usersList = usersData;
        } else if (usersData && Array.isArray(usersData.data)) {
          usersList = usersData.data;
        }
        setSearchUsers(usersList);
      } else {
        setSearchUsers([]);
        showError(response.data.message || '搜索失败');
      }
    } catch (error) {
      console.error('搜索用户失败:', error);
      showError('搜索用户失败');
      setSearchUsers([]);
    } finally {
      setSearching(false);
    }
  }, []);

  // 防抖搜索
  useEffect(() => {
    // 清除之前的定时器
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }

    // 如果搜索词为空，立即清空结果
    if (!searchKeyword.trim()) {
      setSearchUsers([]);
      setHasSearched(false);
      setSearching(false);
      return;
    }

    // 设置新的防抖定时器
    debounceTimerRef.current = setTimeout(() => {
      performSearch(searchKeyword);
    }, 500); // 500ms 防抖延迟

    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, [searchKeyword, performSearch]);

  // 邀请成员
  const handleInvite = async () => {
    if (!selectedUser) {
      showError('请选择要邀请的用户');
      return;
    }

    setInviting(true);
    try {
      const response = await API.post(`/api/team/${teamId}/invite`, {
        user_id: selectedUser.id
      });

      if (response.data.success) {
        showSuccess('邀请发送成功');
        handleClose();
        if (onSuccess) onSuccess();
      } else {
        showError(response.data.message || '邀请失败');
      }
    } catch (error) {
      console.error('邀请成员失败:', error);
      showError('邀请成员失败');
    } finally {
      setInviting(false);
    }
  };

  // 关闭并重置
  const handleClose = () => {
    // 清除防抖计时器
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }
    setSearchKeyword('');
    setSearchUsers([]);
    setSelectedUser(null);
    setHasSearched(false);
    setSearching(false);
    onOpenChange(false);
  };

  // 获取用户名首字母
  const getUserInitial = (username) => {
    return username?.charAt(0).toUpperCase() || 'U';
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <UserPlus className="w-5 h-5" />
            邀请成员
          </DialogTitle>
          <DialogDescription>搜索用户并邀请他们加入团队</DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* 搜索框 */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
            <Input
              placeholder="输入用户名或邮箱进行搜索..."
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
              className="pl-10"
              disabled={inviting}
              autoFocus
            />
          </div>

          {/* 搜索结果 - 固定高度避免抖动 */}
          <div className="h-[240px] flex flex-col">
            {searching ? (
              // 加载状态
              <div className="space-y-2">
                {[...Array(3)].map((_, index) => (
                  <Card key={index}>
                    <CardContent className="p-3">
                      <div className="flex items-center gap-3">
                        <Skeleton className="h-10 w-10 rounded-full" />
                        <div className="flex-1 space-y-2">
                          <Skeleton className="h-4 w-24" />
                          <Skeleton className="h-3 w-32" />
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : searchUsers.length > 0 ? (
              // 搜索结果列表
              <div className="space-y-2 overflow-y-auto flex-1">
                {searchUsers.map((user) => (
                  <Card
                    key={user.id}
                    className={cn(
                      'cursor-pointer transition-all',
                      selectedUser?.id === user.id ? 'border-primary bg-primary/5' : 'hover:bg-muted/50'
                    )}
                    onClick={() => setSelectedUser(user)}
                  >
                    <CardContent className="p-3">
                      <div className="flex items-center gap-3">
                        <Avatar className="h-10 w-10">
                          <AvatarFallback>{getUserInitial(user.username)}</AvatarFallback>
                        </Avatar>
                        <div className="flex-1 min-w-0">
                          <div className="font-medium truncate">{user.username}</div>
                          <div className="text-sm text-muted-foreground truncate">{user.email}</div>
                        </div>
                        {selectedUser?.id === user.id && (
                          <div className="w-5 h-5 rounded-full bg-primary flex items-center justify-center">
                            <div className="w-2 h-2 rounded-full bg-white" />
                          </div>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : hasSearched && searchUsers.length === 0 ? (
              // 无结果提示
              <div className="flex items-center justify-center h-full">
                <Card className="border-dashed w-full">
                  <CardContent className="p-6 text-center text-muted-foreground">未找到匹配的用户</CardContent>
                </Card>
              </div>
            ) : (
              // 初始空状态
              <div className="flex items-center justify-center h-full">
                <div className="text-center text-muted-foreground">
                  <Search className="w-12 h-12 mx-auto mb-3 opacity-20" />
                  <p className="text-sm">输入用户名或邮箱开始搜索</p>
                </div>
              </div>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={inviting}>
            取消
          </Button>
          <Button onClick={handleInvite} disabled={!selectedUser || inviting}>
            {inviting ? '邀请中...' : '发送邀请'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

InviteMemberDialog.propTypes = {
  open: PropTypes.bool.isRequired,
  onOpenChange: PropTypes.func.isRequired,
  teamId: PropTypes.number.isRequired,
  onSuccess: PropTypes.func
};

export default InviteMemberDialog;
