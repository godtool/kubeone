import {post} from "@/plugins/request"

const baseUrl = "/api/v1/roles"

export function searchRoles(pageNum, pageSize, conditions) {
    return post(`${baseUrl}/search?pageNum=${pageNum}&&pageSize=${pageSize}`, conditions)
}