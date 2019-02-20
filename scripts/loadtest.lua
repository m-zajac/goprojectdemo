
math.randomseed(os.time())

languages = {
    "golang",
    "java",
    "javascript",
    "python",
    "scala",
    "php"
}

counts = {3, 5, 6, 10}

randomData = function(dataArray)
    return dataArray[math.random(1, table.getn(dataArray))]
end

request = function()
    wrk.headers["Connection"] = "Keep-Alive"

    return wrk.format("GET", "/bestcontributors/" .. randomData(languages) .. "?count=" .. randomData(counts) .. "&projectsCount=" .. randomData(counts))
end
