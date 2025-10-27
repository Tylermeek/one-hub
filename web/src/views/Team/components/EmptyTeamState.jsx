import PropTypes from 'prop-types';
import { Users } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

const EmptyTeamState = ({ onCreateTeam, className }) => {
    return (
        <Card className={cn('border-dashed', className)}>
            <CardContent className="flex flex-col items-center justify-center py-12 px-6">
                <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center mb-4">
                    <Users className="w-8 h-8 text-muted-foreground" />
                </div>
                <h3 className="text-lg font-semibold text-foreground mb-2">还没有任何团队</h3>
                <p className="text-sm text-muted-foreground text-center mb-6 max-w-sm">创建团队来管理成员和共享额度，让团队协作更高效</p>
                <Button onClick={onCreateTeam} size="lg">
                    <Users className="w-4 h-4 mr-2" />
                    创建第一个团队
                </Button>
            </CardContent>
        </Card>
    );
};

EmptyTeamState.propTypes = {
    onCreateTeam: PropTypes.func.isRequired,
    className: PropTypes.string
};

export default EmptyTeamState;
