import React, { useState, useEffect, useMemo } from 'react';
import PropTypes from 'prop-types';
import {
    flexRender,
    getCoreRowModel,
    getFilteredRowModel,
    getPaginationRowModel,
    getSortedRowModel,
    useReactTable
} from '@tanstack/react-table';
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Eye } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { IconLayoutColumns } from '@tabler/icons-react';

// shadcn components
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardAction } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { DropdownMenu, DropdownMenuContent, DropdownMenuCheckboxItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Drawer, DrawerContent, DrawerDescription, DrawerFooter, DrawerHeader, DrawerTitle, DrawerTrigger } from '@/components/ui/drawer';
import { API } from 'utils/api';
import { showError, renderNumber, renderQuota } from 'utils/common';
import { cn } from '@/lib/utils';

// 格式化时间戳
const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp * 1000);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
};

// 获取类型显示文本
const getTypeText = (type) => {
    const typeMap = {
        1: '文本生成',
        2: '图像生成',
        3: '语音合成',
        4: '语音识别',
        5: '嵌入向量',
        6: '其他'
    };
    return typeMap[type] || '未知';
};

// 获取状态显示
const getStatusDisplay = (status) => {
    if (status === 1) {
        return (
            <Badge variant="default" className="text-xs">
                成功
            </Badge>
        );
    } else if (status === 0) {
        return (
            <Badge variant="destructive" className="text-xs">
                失败
            </Badge>
        );
    } else {
        return (
            <Badge variant="secondary" className="text-xs">
                未知
            </Badge>
        );
    }
};

// 表格列定义
const createColumns = () => [
    {
        accessorKey: 'created_at',
        header: '时间',
        cell: ({ getValue }) => formatTimestamp(getValue())
    },
    {
        accessorKey: 'type',
        header: '类型',
        cell: ({ getValue }) => getTypeText(getValue())
    },
    {
        accessorKey: 'model_name',
        header: '模型',
        cell: ({ getValue }) => <span className="font-mono text-sm">{getValue()}</span>
    },
    {
        accessorKey: 'token_name',
        header: 'Token',
        cell: ({ getValue }) => <span className="text-sm text-muted-foreground">{getValue()}</span>
    },
    {
        accessorKey: 'quota',
        header: '消费',
        cell: ({ getValue }) => renderQuota(getValue())
    },
    {
        id: 'tokens',
        header: 'Token用量',
        cell: ({ row }) => {
            const promptTokens = row.original.prompt_tokens || 0;
            const completionTokens = row.original.completion_tokens || 0;
            const total = promptTokens + completionTokens;
            return (
                <div className="text-sm">
                    <div className="font-medium">{renderNumber(total)}</div>
                    <div className="text-xs text-muted-foreground">
                        P:{renderNumber(promptTokens)} + C:{renderNumber(completionTokens)}
                    </div>
                </div>
            );
        }
    },
    {
        accessorKey: 'status',
        header: '状态',
        cell: ({ getValue }) => getStatusDisplay(getValue())
    },
    {
        id: 'actions',
        header: '操作',
        cell: ({ row }) => (
            <Drawer>
                <DrawerTrigger asChild>
                    <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                        <Eye className="h-4 w-4" />
                        <span className="sr-only">查看详情</span>
                    </Button>
                </DrawerTrigger>
                <DrawerContent>
                    <DrawerHeader>
                        <DrawerTitle>API 调用详情</DrawerTitle>
                        <DrawerDescription>查看详细的 API 调用信息</DrawerDescription>
                    </DrawerHeader>
                    <div className="px-4 pb-4 space-y-4">
                        <div className="grid grid-cols-2 gap-4 text-sm">
                            <div>
                                <span className="text-muted-foreground">时间:</span>
                                <div className="font-medium">{formatTimestamp(row.original.created_at)}</div>
                            </div>
                            <div>
                                <span className="text-muted-foreground">类型:</span>
                                <div className="font-medium">{getTypeText(row.original.type)}</div>
                            </div>
                            <div>
                                <span className="text-muted-foreground">模型:</span>
                                <div className="font-mono">{row.original.model_name}</div>
                            </div>
                            <div>
                                <span className="text-muted-foreground">Token:</span>
                                <div className="font-medium">{row.original.token_name}</div>
                            </div>
                            <div>
                                <span className="text-muted-foreground">消费:</span>
                                <div className="font-medium">{renderQuota(row.original.quota)}</div>
                            </div>
                            <div>
                                <span className="text-muted-foreground">状态:</span>
                                <div>{getStatusDisplay(row.original.status)}</div>
                            </div>
                        </div>
                        <div className="space-y-2">
                            <span className="text-muted-foreground text-sm">Token 使用详情:</span>
                            <div className="grid grid-cols-2 gap-4 text-sm">
                                <div>
                                    <span className="text-muted-foreground">输入 Token:</span>
                                    <div className="font-medium">{renderNumber(row.original.prompt_tokens || 0)}</div>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">输出 Token:</span>
                                    <div className="font-medium">{renderNumber(row.original.completion_tokens || 0)}</div>
                                </div>
                            </div>
                        </div>
                    </div>
                    <DrawerFooter>
                        <Button variant="outline">关闭</Button>
                    </DrawerFooter>
                </DrawerContent>
            </Drawer>
        )
    }
];

// TODO 渲染 逻辑 请迁移 保证 和 日志页面 列表 的一样，统一

const RecentLogsTable = ({ className = '' }) => {
    const { t } = useTranslation();
    const [data, setData] = useState([]);
    const [isLoading, setIsLoading] = useState(true);
    const [timeRange, setTimeRange] = useState('7d');
    const [pagination, setPagination] = useState({
        pageIndex: 0,
        pageSize: 10
    });
    const [columnVisibility, setColumnVisibility] = useState({});
    const [total, setTotal] = useState(0);

    // 获取数据
    const fetchData = async () => {
        setIsLoading(true);
        try {
            const response = await API.get('/api/user/dashboard/recent-logs', {
                params: {
                    page: pagination.pageIndex + 1,
                    page_size: pagination.pageSize,
                    time_range: timeRange
                }
            });

            const { success, message, data: responseData } = response.data;
            if (success && responseData) {
                setData(responseData.items || []);
                setTotal(responseData.total || 0);
            } else {
                showError(message);
            }
        } catch (error) {
            console.error('Error fetching recent logs:', error);
            showError('获取日志数据失败');
        } finally {
            setIsLoading(false);
        }
    };

    // 监听参数变化
    useEffect(() => {
        fetchData();
    }, [pagination.pageIndex, pagination.pageSize, timeRange]);

    // 创建表格列
    const columns = useMemo(() => createColumns(t), [t]);

    // 创建表格实例
    const table = useReactTable({
        data,
        columns,
        state: {
            pagination,
            columnVisibility
        },
        onPaginationChange: setPagination,
        onColumnVisibilityChange: setColumnVisibility,
        getCoreRowModel: getCoreRowModel(),
        getFilteredRowModel: getFilteredRowModel(),
        getPaginationRowModel: getPaginationRowModel(),
        getSortedRowModel: getSortedRowModel(),
        manualPagination: true,
        pageCount: Math.ceil(total / pagination.pageSize)
    });

    return (
        <Card className={cn('@container/card h-auto', className)}>
            <CardHeader>
                <CardTitle>最近调用记录</CardTitle>
                <CardDescription>
                    <span className="hidden @[540px]/card:block">
                        最近 {timeRange === '7d' ? '7' : timeRange === '30d' ? '30' : '90'} 天的 API 调用记录
                    </span>
                    <span className="@[540px]/card:hidden">最近 {timeRange === '7d' ? '7' : timeRange === '30d' ? '30' : '90'} 天</span>
                </CardDescription>
                <CardAction>
                    <div className="flex items-center gap-2">
                        <Select value={timeRange} onValueChange={setTimeRange}>
                            <SelectTrigger className="w-32" size="sm">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="7d">最近 7 天</SelectItem>
                                <SelectItem value="30d">最近 30 天</SelectItem>
                                <SelectItem value="90d">最近 3 个月</SelectItem>
                            </SelectContent>
                        </Select>
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button variant="outline" size="sm">
                                    <IconLayoutColumns className="h-4 w-4" />
                                    <span className="hidden lg:inline ml-2">自定义列</span>
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end" className="w-48">
                                {table
                                    .getAllColumns()
                                    .filter((column) => column.getCanHide())
                                    .map((column) => {
                                        return (
                                            <DropdownMenuCheckboxItem
                                                key={column.id}
                                                className="capitalize"
                                                checked={column.getIsVisible()}
                                                onCheckedChange={(value) => column.toggleVisibility(!!value)}
                                            >
                                                {column.id}
                                            </DropdownMenuCheckboxItem>
                                        );
                                    })}
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </div>
                </CardAction>
            </CardHeader>
            <CardContent>
                {isLoading ? (
                    <div className="space-y-4">
                        <Skeleton className="h-10 w-full" />
                        {[...Array(5)].map((_, i) => (
                            <Skeleton key={i} className="h-12 w-full" />
                        ))}
                    </div>
                ) : (
                    <div className="space-y-4">
                        <div className="rounded-md border">
                            <Table>
                                <TableHeader>
                                    {table.getHeaderGroups().map((headerGroup) => (
                                        <TableRow key={headerGroup.id}>
                                            {headerGroup.headers.map((header) => {
                                                return (
                                                    <TableHead key={header.id}>
                                                        {header.isPlaceholder
                                                            ? null
                                                            : flexRender(header.column.columnDef.header, header.getContext())}
                                                    </TableHead>
                                                );
                                            })}
                                        </TableRow>
                                    ))}
                                </TableHeader>
                                <TableBody>
                                    {table.getRowModel().rows?.length ? (
                                        table.getRowModel().rows.map((row) => (
                                            <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
                                                {row.getVisibleCells().map((cell) => (
                                                    <TableCell key={cell.id}>
                                                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                                    </TableCell>
                                                ))}
                                            </TableRow>
                                        ))
                                    ) : (
                                        <TableRow>
                                            <TableCell colSpan={columns.length} className="h-24 text-center">
                                                暂无数据
                                            </TableCell>
                                        </TableRow>
                                    )}
                                </TableBody>
                            </Table>
                        </div>

                        {/* 分页器 */}
                        <div className="flex items-center justify-between">
                            <div className="text-sm text-muted-foreground">共 {total} 条记录</div>
                            <div className="flex items-center gap-2">
                                <div className="flex items-center gap-2">
                                    <span className="text-sm">每页显示</span>
                                    <Select
                                        value={`${table.getState().pagination.pageSize}`}
                                        onValueChange={(value) => {
                                            table.setPageSize(Number(value));
                                        }}
                                    >
                                        <SelectTrigger className="h-8 w-16">
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {[10, 20, 30, 40, 50].map((pageSize) => (
                                                <SelectItem key={pageSize} value={`${pageSize}`}>
                                                    {pageSize}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="flex items-center gap-1">
                                    <Button
                                        variant="outline"
                                        className="h-8 w-8 p-0"
                                        onClick={() => table.setPageIndex(0)}
                                        disabled={!table.getCanPreviousPage()}
                                    >
                                        <ChevronsLeft className="h-4 w-4" />
                                    </Button>
                                    <Button
                                        variant="outline"
                                        className="h-8 w-8 p-0"
                                        onClick={() => table.previousPage()}
                                        disabled={!table.getCanPreviousPage()}
                                    >
                                        <ChevronLeft className="h-4 w-4" />
                                    </Button>
                                    <div className="flex items-center justify-center text-sm font-medium w-20">
                                        第 {table.getState().pagination.pageIndex + 1} 页
                                    </div>
                                    <Button
                                        variant="outline"
                                        className="h-8 w-8 p-0"
                                        onClick={() => table.nextPage()}
                                        disabled={!table.getCanNextPage()}
                                    >
                                        <ChevronRight className="h-4 w-4" />
                                    </Button>
                                    <Button
                                        variant="outline"
                                        className="h-8 w-8 p-0"
                                        onClick={() => table.setPageIndex(table.getPageCount() - 1)}
                                        disabled={!table.getCanNextPage()}
                                    >
                                        <ChevronsRight className="h-4 w-4" />
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </CardContent>
        </Card>
    );
};

RecentLogsTable.propTypes = {
    className: PropTypes.string
};

export default RecentLogsTable;
