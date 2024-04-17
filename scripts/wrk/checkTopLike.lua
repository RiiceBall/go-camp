response = function(status, headers, body)
    local isOrdered = checkLikeOrder(body)
    if not isOrdered then
        -- 如果没有打印则表示正常
        print("返回的点赞数据不是从最大到最小的")
    end
end

-- 检查点赞数是否有序
function checkLikeOrder(data)
    local like_cnt_values = {}
    for like_cnt in data:gmatch('"like_cnt":(%d+)') do
        table.insert(like_cnt_values, tonumber(like_cnt))
    end
    for i = 1, #like_cnt_values - 1 do
        if like_cnt_values[i] < like_cnt_values[i + 1] then
            return false
        end
    end
    return true
end
