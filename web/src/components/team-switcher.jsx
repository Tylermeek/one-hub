'use client';

import { useEffect } from 'react';
import { ChevronsUpDown, Home, Users } from 'lucide-react';
import { useSelector } from 'react-redux';

import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar } from '@/components/ui/sidebar';
import useContext from '@/hooks/useContext';

export function TeamSwitcher() {
    const { isMobile } = useSidebar();
    const { currentContext, userTeams } = useSelector((state) => state.context);
    const { user } = useSelector((state) => state.account);
    const { switchContext, loadUserTeams } = useContext();

    // 在组件挂载时加载用户团队数据
    useEffect(() => {
        if (user && userTeams.length === 0) {
            console.log('TeamSwitcher: Loading user teams for user:', user.username);
            loadUserTeams();
        }
    }, [user, userTeams.length, loadUserTeams]);

    // 每次打开菜单时也刷新团队列表（参考旧版逻辑）
    const handleMenuOpen = () => {
        if (user) {
            console.log('TeamSwitcher: Refreshing teams on menu open');
            loadUserTeams();
        }
    };

    // 调试信息
    useEffect(() => {
        console.log('TeamSwitcher Debug:', {
            user: user ? user.username : 'No user',
            userTeams: userTeams,
            currentContext: currentContext
        });
    }, [user, userTeams, currentContext]);

    // 如果用户未登录，不显示切换器
    if (!user) {
        console.log('TeamSwitcher: Not showing - no user');
        return null;
    }

    const handleSwitchContext = (context) => {
        switchContext(context);
    };

    // 获取上下文图标
    const getContextIcon = (type) => {
        return type === 'user' ? Home : Users;
    };

    // 获取上下文颜色类
    const getContextColorClass = (type) => {
        return type === 'user' ? 'bg-primary text-primary-foreground' : 'bg-secondary text-secondary-foreground';
    };

    const ContextIcon = getContextIcon(currentContext.type);

    return (
        <SidebarMenu>
            <SidebarMenuItem>
                <DropdownMenu
                    onOpenChange={(open) => {
                        if (open) {
                            handleMenuOpen();
                        }
                    }}
                >
                    <DropdownMenuTrigger asChild>
                        <SidebarMenuButton
                            size="lg"
                            className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                        >
                            <div
                                className={`flex aspect-square size-8 items-center justify-center rounded-lg ${getContextColorClass(currentContext.type)}`}
                            >
                                <ContextIcon className="size-4" />
                            </div>
                            <div className="grid flex-1 text-left text-sm leading-tight">
                                <span className="truncate font-medium">{currentContext.name}</span>
                                <span className="truncate text-xs">{currentContext.type === 'user' ? '个人空间' : '团队空间'}</span>
                            </div>
                            <ChevronsUpDown className="ml-auto" />
                        </SidebarMenuButton>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent
                        className="w-56 min-w-56 rounded-lg z-50"
                        align="start"
                        side={isMobile ? 'bottom' : 'right'}
                        sideOffset={4}
                    >
                        <DropdownMenuLabel className="text-muted-foreground text-xs">空间切换</DropdownMenuLabel>

                        {/* 个人空间 */}
                        <DropdownMenuItem
                            onClick={() =>
                                handleSwitchContext({
                                    type: 'user',
                                    id: user.id,
                                    name: '个人空间'
                                })
                            }
                            className="gap-2 p-2"
                        >
                            <div className="flex size-6 items-center justify-center rounded-md border bg-primary text-primary-foreground">
                                <Home className="size-3.5 shrink-0" />
                            </div>
                            <div className="flex flex-col">
                                <span className="font-medium">个人空间</span>
                                <span className="text-xs text-muted-foreground">个人资源和设置</span>
                            </div>
                        </DropdownMenuItem>

                        {/* 团队列表 */}
                        {userTeams.length > 0 && (
                            <>
                                <DropdownMenuSeparator />
                                {userTeams.map((team) => (
                                    <DropdownMenuItem
                                        key={team.id}
                                        onClick={() =>
                                            handleSwitchContext({
                                                type: 'team',
                                                id: team.id,
                                                name: team.name
                                            })
                                        }
                                        className="gap-2 p-2"
                                    >
                                        <div className="flex size-6 items-center justify-center rounded-md border bg-secondary text-secondary-foreground">
                                            <Users className="size-3.5 shrink-0" />
                                        </div>
                                        <div className="flex flex-col">
                                            <span className="font-medium">{team.name}</span>
                                            <span className="text-xs text-muted-foreground">团队空间</span>
                                        </div>
                                    </DropdownMenuItem>
                                ))}
                            </>
                        )}

                        {/* 创建团队按钮 */}
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                            onClick={() => {
                                // TODO: 实现创建团队逻辑
                                console.log('Create team clicked');
                            }}
                            className="gap-2 p-2 text-muted-foreground"
                        >
                            <div className="flex size-6 items-center justify-center rounded-md border border-dashed">
                                <Users className="size-3.5 shrink-0" />
                            </div>
                            <div className="flex flex-col">
                                <span className="font-medium">创建团队</span>
                                <span className="text-xs text-muted-foreground">新建团队空间</span>
                            </div>
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            </SidebarMenuItem>
        </SidebarMenu>
    );
}
