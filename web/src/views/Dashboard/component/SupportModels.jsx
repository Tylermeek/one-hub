import { useState, useEffect } from 'react';
import { API } from 'utils/api';
import { showError, copy } from 'utils/common';
import { useTranslation } from 'react-i18next';
import { ChevronDown, ChevronUp, Search } from 'lucide-react';
import { useSelector } from 'react-redux';
import IconWrapper from 'ui-component/IconWrapper';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

const SupportModels = () => {
    const [modelList, setModelList] = useState([]);
    const [expanded, setExpanded] = useState(false);
    const [searchTerm, setSearchTerm] = useState('');
    const { t } = useTranslation();
    const ownedby = useSelector((state) => state.siteInfo?.ownedby);

    const fetchModels = async () => {
        try {
            const res = await API.get(`/api/available_model`);
            const { data, success } = res.data;
            if (!success) return;

            const modelGroup = Object.entries(data).reduce((acc, [modelId, modelInfo]) => {
                const { owned_by } = modelInfo;
                if (!acc[owned_by]) {
                    acc[owned_by] = [];
                }
                acc[owned_by].push(modelId);
                return acc;
            }, {});

            Object.values(modelGroup).forEach((models) => models.sort());

            const sortedModelGroup = Object.keys(modelGroup)
                .sort()
                .reduce((acc, key) => {
                    acc[key] = modelGroup[key];
                    return acc;
                }, {});

            setModelList(sortedModelGroup);
        } catch (error) {
            showError(error.message);
        }
    };

    useEffect(() => {
        fetchModels();
    }, []);

    const getIconByName = (name) => {
        const owner = ownedby.find((item) => item.name === name);
        return owner?.icon;
    };

    // 过滤模型列表
    const safeModelList = modelList || {};
    const filteredModelList = Object.entries(safeModelList).reduce((acc, [provider, models]) => {
        const safeModels = models || [];
        const filteredModels = safeModels.filter(
            (model) => model.toLowerCase().includes(searchTerm.toLowerCase()) || provider.toLowerCase().includes(searchTerm.toLowerCase())
        );
        if (filteredModels.length > 0) {
            acc[provider] = filteredModels;
        }
        return acc;
    }, {});

    return (
        <Card>
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <span>{t('dashboard_index.model_price')}</span>
                    <Badge variant="secondary" className="ml-2">
                        {Object.values(safeModelList).flat().length} 个模型
                    </Badge>
                </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
                {/* 搜索框 */}
                <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
                    <Input
                        placeholder="搜索模型或提供商..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="pl-10"
                    />
                </div>

                {/* 预览模式 */}
                {!expanded && (
                    <div className="space-y-3">
                        {Object.entries(filteredModelList)
                            .slice(0, 2)
                            .map(([provider, models]) => (
                                <div key={provider} className="space-y-2">
                                    <div className="flex items-center gap-2">
                                        <IconWrapper url={getIconByName(provider)} />
                                        <span className="text-sm font-medium text-muted-foreground">{provider}</span>
                                        <Badge variant="outline" className="text-xs">
                                            {models.length} 个模型
                                        </Badge>
                                    </div>
                                    <div className="flex flex-wrap gap-1.5">
                                        {models.slice(0, 4).map((model) => (
                                            <Badge
                                                key={model}
                                                variant="secondary"
                                                className="cursor-pointer hover:bg-primary/20 transition-colors text-xs"
                                                onClick={() => copy(model, t('dashboard_index.model_name'))}
                                            >
                                                {model}
                                            </Badge>
                                        ))}
                                        {models.length > 4 && (
                                            <Badge variant="outline" className="text-xs">
                                                +{models.length - 4} 更多
                                            </Badge>
                                        )}
                                    </div>
                                </div>
                            ))}
                    </div>
                )}

                {/* 展开模式 */}
                {expanded && (
                    <div className="space-y-4">
                        {Object.entries(filteredModelList).map(([provider, models]) => (
                            <div key={provider} className="space-y-3">
                                <div className="flex items-center gap-2">
                                    <IconWrapper url={getIconByName(provider)} />
                                    <span className="text-sm font-semibold">{provider}</span>
                                    <Badge variant="outline" className="text-xs">
                                        {models.length} 个模型
                                    </Badge>
                                </div>
                                <div className="flex flex-wrap gap-2 pl-6">
                                    {models.map((model) => (
                                        <Badge
                                            key={model}
                                            variant="secondary"
                                            className="cursor-pointer hover:bg-primary/20 transition-colors"
                                            onClick={() => copy(model, t('dashboard_index.model_name'))}
                                        >
                                            {model}
                                        </Badge>
                                    ))}
                                </div>
                            </div>
                        ))}
                        {Object.keys(filteredModelList).length === 0 && (
                            <div className="text-center py-8 text-muted-foreground">
                                <Search className="w-8 h-8 mx-auto mb-2 opacity-50" />
                                <p>未找到匹配的模型</p>
                            </div>
                        )}
                    </div>
                )}

                {/* 展开/收起按钮 */}
                <div className="flex justify-center pt-2">
                    <TooltipProvider>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => setExpanded(!expanded)}
                                    className="text-muted-foreground hover:text-foreground"
                                >
                                    {expanded ? (
                                        <>
                                            <ChevronUp className="w-4 h-4 mr-1" />
                                            收起
                                        </>
                                    ) : (
                                        <>
                                            <ChevronDown className="w-4 h-4 mr-1" />
                                            查看全部
                                        </>
                                    )}
                                </Button>
                            </TooltipTrigger>
                            <TooltipContent>
                                <p>{expanded ? '收起模型列表' : '展开查看所有模型'}</p>
                            </TooltipContent>
                        </Tooltip>
                    </TooltipProvider>
                </div>
            </CardContent>
        </Card>
    );
};

export default SupportModels;
