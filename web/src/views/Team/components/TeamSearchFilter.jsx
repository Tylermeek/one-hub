import PropTypes from 'prop-types';
import { Search } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { cn } from '@/lib/utils';

const TeamSearchFilter = ({
    searchTerm,
    onSearchChange,
    statusFilter,
    onStatusFilterChange,
    roleFilter,
    onRoleFilterChange,
    className
}) => {
    return (
        <div className={cn('flex flex-col sm:flex-row gap-3', className)}>
            {/* 搜索框 */}
            <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
                <Input
                    placeholder="搜索团队名称..."
                    value={searchTerm}
                    onChange={(e) => onSearchChange(e.target.value)}
                    className="pl-10"
                />
            </div>

            {/* 状态筛选 */}
            <Select value={statusFilter} onValueChange={onStatusFilterChange}>
                <SelectTrigger className="w-full sm:w-32">
                    <SelectValue placeholder="状态" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="all">全部状态</SelectItem>
                    <SelectItem value="active">正常</SelectItem>
                    <SelectItem value="disabled">禁用</SelectItem>
                </SelectContent>
            </Select>

            {/* 角色筛选 */}
            <Select value={roleFilter} onValueChange={onRoleFilterChange}>
                <SelectTrigger className="w-full sm:w-32">
                    <SelectValue placeholder="角色" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="all">全部角色</SelectItem>
                    <SelectItem value="owner">Owner</SelectItem>
                    <SelectItem value="admin">Admin</SelectItem>
                    <SelectItem value="member">Member</SelectItem>
                </SelectContent>
            </Select>
        </div>
    );
};

TeamSearchFilter.propTypes = {
    searchTerm: PropTypes.string.isRequired,
    onSearchChange: PropTypes.func.isRequired,
    statusFilter: PropTypes.string.isRequired,
    onStatusFilterChange: PropTypes.func.isRequired,
    roleFilter: PropTypes.string.isRequired,
    onRoleFilterChange: PropTypes.func.isRequired,
    className: PropTypes.string
};

export default TeamSearchFilter;
