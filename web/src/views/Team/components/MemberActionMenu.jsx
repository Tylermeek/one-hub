import PropTypes from 'prop-types';
import { MoreVertical, Settings, UserX, Shield, ShieldOff } from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { Button } from '@/components/ui/button';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle
} from '@/components/ui/alert-dialog';
import { useState } from 'react';
import { canManageSpecificMember, ROLE } from 'utils/teamPermissions';

const MemberActionMenu = ({ member, currentUserRole, onSetQuota, onChangeRole, onRemove }) => {
    const [showRemoveDialog, setShowRemoveDialog] = useState(false);
    const [showRoleDialog, setShowRoleDialog] = useState(false);
    const [targetRole, setTargetRole] = useState(null);

    // 检查是否可以管理这个特定成员
    const canManageThisMember = canManageSpecificMember(currentUserRole, member.role);

    if (!canManageThisMember) {
        return null;
    }

    const handleChangeRole = (newRole) => {
        setTargetRole(newRole);
        setShowRoleDialog(true);
    };

    const confirmRoleChange = () => {
        if (onChangeRole && targetRole !== null) {
            onChangeRole(member.user_id, targetRole);
        }
        setShowRoleDialog(false);
        setTargetRole(null);
    };

    const confirmRemove = () => {
        if (onRemove) {
            onRemove(member.user_id);
        }
        setShowRemoveDialog(false);
    };

    return (
        <>
            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                        <MoreVertical className="h-4 w-4" />
                        <span className="sr-only">操作菜单</span>
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-48">
                    <DropdownMenuLabel>成员操作</DropdownMenuLabel>
                    <DropdownMenuSeparator />

                    <DropdownMenuItem onClick={() => onSetQuota && onSetQuota(member)}>
                        <Settings className="mr-2 h-4 w-4" />
                        设置额度限制
                    </DropdownMenuItem>

                    {/* Member 可以被提升为 Admin */}
                    {member.role === ROLE.MEMBER && (
                        <DropdownMenuItem onClick={() => handleChangeRole(ROLE.ADMIN)}>
                            <Shield className="mr-2 h-4 w-4" />
                            设为管理员
                        </DropdownMenuItem>
                    )}

                    {/* Admin 可以被降为 Member（仅 Owner 可操作）*/}
                    {member.role === ROLE.ADMIN && currentUserRole === ROLE.OWNER && (
                        <DropdownMenuItem onClick={() => handleChangeRole(ROLE.MEMBER)}>
                            <ShieldOff className="mr-2 h-4 w-4" />
                            设为普通成员
                        </DropdownMenuItem>
                    )}

                    {/* 移除成员 - Owner 不能被移除 */}
                    {member.role !== ROLE.OWNER && (
                        <>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setShowRemoveDialog(true)}>
                                <UserX className="mr-2 h-4 w-4" />
                                移除成员
                            </DropdownMenuItem>
                        </>
                    )}
                </DropdownMenuContent>
            </DropdownMenu>

            {/* 移除成员确认对话框 */}
            <AlertDialog open={showRemoveDialog} onOpenChange={setShowRemoveDialog}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>确认移除成员？</AlertDialogTitle>
                        <AlertDialogDescription>
                            您确定要从团队中移除 <span className="font-medium">{member.user?.username}</span> 吗？ 此操作无法撤销。
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={confirmRemove}
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        >
                            确认移除
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>

            {/* 更改角色确认对话框 */}
            <AlertDialog open={showRoleDialog} onOpenChange={setShowRoleDialog}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>确认更改角色？</AlertDialogTitle>
                        <AlertDialogDescription>
                            您确定要将 <span className="font-medium">{member.user?.username}</span> 的角色更改为
                            <span className="font-medium">{targetRole === ROLE.ADMIN ? ' 管理员' : ' 普通成员'}</span> 吗？
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction onClick={confirmRoleChange}>确认更改</AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </>
    );
};

MemberActionMenu.propTypes = {
    member: PropTypes.shape({
        user_id: PropTypes.number.isRequired,
        role: PropTypes.number.isRequired,
        user: PropTypes.shape({
            username: PropTypes.string
        })
    }).isRequired,
    currentUserRole: PropTypes.number.isRequired,
    onSetQuota: PropTypes.func,
    onChangeRole: PropTypes.func,
    onRemove: PropTypes.func
};

export default MemberActionMenu;
