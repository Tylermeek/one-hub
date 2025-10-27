import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { Download, FileSpreadsheet, FileText, ChevronDown, ChevronUp, ArrowUpDown } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { cn } from '@/lib/utils';
import { renderQuota, showSuccess, showError } from 'utils/common';
import { canExportData } from 'utils/teamPermissions';

/**
 * 导出 CSV
 */
const exportToCSV = (data, filename) => {
    if (!data || data.length === 0) {
        showError('No data to export');
        return;
    }

    const headers = Object.keys(data[0]);
    const csvContent = [headers.join(','), ...data.map((row) => headers.map((header) => JSON.stringify(row[header] || '')).join(','))].join(
        '\n'
    );

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = `${filename}.csv`;
    link.click();

    showSuccess('Exported successfully');
};

/**
 * 数据导出表格组件
 */
function DataExportTable({ data, userRole, isLoading, className }) {
    const { t } = useTranslation();
    const [sortField, setSortField] = useState('created_at');
    const [sortOrder, setSortOrder] = useState('desc');
    const [expandedRows, setExpandedRows] = useState(new Set());

    const canExport = canExportData(userRole);

    /**
     * 排序处理
     */
    const handleSort = (field) => {
        if (sortField === field) {
            setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortOrder('desc');
        }
    };

    /**
     * 展开/收起行
     */
    const toggleExpand = (id) => {
        const newExpanded = new Set(expandedRows);
        if (newExpanded.has(id)) {
            newExpanded.delete(id);
        } else {
            newExpanded.add(id);
        }
        setExpandedRows(newExpanded);
    };

    /**
     * 导出数据
     */
    const handleExport = (format) => {
        if (!canExport) {
            showError(t('team_analytics.no_export_permission'));
            return;
        }

        const filename = `team_data_${new Date().toISOString().split('T')[0]}`;

        if (format === 'csv') {
            exportToCSV(data, filename);
        } else if (format === 'json') {
            const blob = new Blob([JSON.stringify(data, null, 2)], {
                type: 'application/json'
            });
            const link = document.createElement('a');
            link.href = URL.createObjectURL(blob);
            link.download = `${filename}.json`;
            link.click();
            showSuccess('Exported successfully');
        }
    };

    /**
     * 排序数据
     */
    const sortedData = React.useMemo(() => {
        if (!data) return [];

        return [...data].sort((a, b) => {
            const aVal = a[sortField];
            const bVal = b[sortField];

            if (aVal === bVal) return 0;

            const compareResult = aVal > bVal ? 1 : -1;
            return sortOrder === 'asc' ? compareResult : -compareResult;
        });
    }, [data, sortField, sortOrder]);

    /**
     * 渲染表头排序图标
     */
    const renderSortIcon = (field) => {
        if (sortField !== field) {
            return <ArrowUpDown className="ml-2 h-4 w-4 text-muted-foreground" />;
        }
        return sortOrder === 'asc' ? <ChevronUp className="ml-2 h-4 w-4" /> : <ChevronDown className="ml-2 h-4 w-4" />;
    };

    if (isLoading) {
        return (
            <Card className={className}>
                <CardHeader>
                    <CardTitle>{t('team_analytics.detailed_data')}</CardTitle>
                </CardHeader>
                <CardContent>
                    <Skeleton className="h-96 w-full" />
                </CardContent>
            </Card>
        );
    }

    return (
        <Card className={className}>
            <CardHeader>
                <div className="flex items-center justify-between">
                    <CardTitle>{t('team_analytics.detailed_data')}</CardTitle>

                    {canExport && (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button variant="outline" size="sm">
                                    <Download className="mr-2 h-4 w-4" />
                                    {t('common.export')}
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                                <DropdownMenuItem onClick={() => handleExport('csv')}>
                                    <FileSpreadsheet className="mr-2 h-4 w-4" />
                                    {t('common.export_csv')}
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => handleExport('json')}>
                                    <FileText className="mr-2 h-4 w-4" />
                                    {t('common.export_json')}
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    )}
                </div>
            </CardHeader>
            <CardContent>
                {!sortedData || sortedData.length === 0 ? (
                    <div className="text-center py-12 text-muted-foreground">{t('common.no_data')}</div>
                ) : (
                    <div className="rounded-md border">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead className="w-12"></TableHead>
                                    <TableHead>
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            className="h-8 p-0 font-medium"
                                            onClick={() => handleSort('created_at')}
                                        >
                                            {t('common.time')}
                                            {renderSortIcon('created_at')}
                                        </Button>
                                    </TableHead>
                                    <TableHead>
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            className="h-8 p-0 font-medium"
                                            onClick={() => handleSort('user_name')}
                                        >
                                            {t('common.user')}
                                            {renderSortIcon('user_name')}
                                        </Button>
                                    </TableHead>
                                    <TableHead>
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            className="h-8 p-0 font-medium"
                                            onClick={() => handleSort('model_name')}
                                        >
                                            {t('common.model')}
                                            {renderSortIcon('model_name')}
                                        </Button>
                                    </TableHead>
                                    <TableHead className="text-right">
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            className="h-8 p-0 font-medium"
                                            onClick={() => handleSort('quota')}
                                        >
                                            {t('common.quota')}
                                            {renderSortIcon('quota')}
                                        </Button>
                                    </TableHead>
                                    <TableHead>{t('common.status')}</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {sortedData.map((row) => (
                                    <React.Fragment key={row.id}>
                                        <TableRow className="hover:bg-muted/50">
                                            <TableCell>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    className="h-6 w-6 p-0"
                                                    onClick={() => toggleExpand(row.id)}
                                                >
                                                    {expandedRows.has(row.id) ? (
                                                        <ChevronUp className="h-4 w-4" />
                                                    ) : (
                                                        <ChevronDown className="h-4 w-4" />
                                                    )}
                                                </Button>
                                            </TableCell>
                                            <TableCell className="text-sm text-muted-foreground">
                                                {new Date(row.created_at * 1000).toLocaleString()}
                                            </TableCell>
                                            <TableCell className="font-medium">{row.user_name}</TableCell>
                                            <TableCell>
                                                <Badge variant="outline">{row.model_name}</Badge>
                                            </TableCell>
                                            <TableCell className="text-right font-mono">{renderQuota(row.quota)}</TableCell>
                                            <TableCell>
                                                <Badge variant={row.status === 'success' ? 'default' : 'destructive'}>{row.status}</Badge>
                                            </TableCell>
                                        </TableRow>

                                        {/* 展开的详细信息 */}
                                        {expandedRows.has(row.id) && (
                                            <TableRow>
                                                <TableCell colSpan={6} className="bg-muted/30">
                                                    <div className="p-4 space-y-2 text-sm">
                                                        <div className="grid grid-cols-2 gap-4">
                                                            <div>
                                                                <span className="text-muted-foreground">Request ID:</span>
                                                                <span className="ml-2 font-mono">{row.id}</span>
                                                            </div>
                                                            <div>
                                                                <span className="text-muted-foreground">Channel:</span>
                                                                <span className="ml-2">{row.channel_name}</span>
                                                            </div>
                                                            <div>
                                                                <span className="text-muted-foreground">Tokens:</span>
                                                                <span className="ml-2">{row.tokens || 'N/A'}</span>
                                                            </div>
                                                            <div>
                                                                <span className="text-muted-foreground">Duration:</span>
                                                                <span className="ml-2">{row.duration || 'N/A'}ms</span>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </TableCell>
                                            </TableRow>
                                        )}
                                    </React.Fragment>
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                )}
            </CardContent>
        </Card>
    );
}

DataExportTable.propTypes = {
    data: PropTypes.arrayOf(PropTypes.object),
    userRole: PropTypes.number.isRequired,
    isLoading: PropTypes.bool,
    className: PropTypes.string
};

export default DataExportTable;
