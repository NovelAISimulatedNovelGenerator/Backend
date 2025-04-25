# -*- coding: utf-8 -*-
"""
test_save_get.py
获取保存接口（/api/save/get）API单元测试

- 覆盖正常获取、获取不存在、缺失参数、鉴权失败等情况
- 需先启动后端服务
- 依赖 pytest fixture auto_register_and_login，自动注册并登录测试用户
"""
import requests
import pytest

BASE_URL = "http://127.0.0.1:8888"
CREATE_SAVE_URL = f"{BASE_URL}/api/save/create"
GET_SAVE_URL = f"{BASE_URL}/api/save/get"

class TestSaveGet:
    """
    /api/save/get 获取接口测试
    """
    def create_save_and_get_id(self, token: str) -> str:
        """
        辅助方法：创建存档并返回 save_id
        """
        payload = {
            "save_name": "Get测试存档",
            "save_data": "内容",
            "save_type": "novel",
            "save_description": "章节2"
        }
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.post(CREATE_SAVE_URL, json=payload, headers=headers)
        assert resp.status_code == 200
        return resp.json()["save_id"]

    def test_get_save_success(self, auto_register_and_login):
        """
        正常获取已存在存档
        """
        token = auto_register_and_login
        save_id = self.create_save_and_get_id(token)
        params = {"save_id": save_id}
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.get(GET_SAVE_URL, params=params, headers=headers)
        print(f"[test_get_save_success] status={resp.status_code}, resp={resp.text}")
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200
        assert data["message"] == "获取成功"
        assert data["save"] is not None
        assert data["save"]["save_id"] == save_id

    def test_get_save_not_found(self, auto_register_and_login):
        """
        获取不存在的存档，应返回404
        """
        token = auto_register_and_login
        params = {"save_id": "999999999999"}
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.get(GET_SAVE_URL, params=params, headers=headers)
        print(f"[test_get_save_not_found] status={resp.status_code}, resp={resp.text}")
        assert resp.status_code == 404
        data = resp.json()
        assert data["code"] == 404
        assert data["message"] == "保存项不存在"

    def test_get_save_missing_params(self, auto_register_and_login):
        """
        缺少 save_id 参数，应返回400
        """
        token = auto_register_and_login
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.get(GET_SAVE_URL, headers=headers)
        print(f"[test_get_save_missing_params] status={resp.status_code}, resp={resp.text}")
        assert resp.status_code == 400

    def test_get_save_unauthorized(self):
        """
        未携带token，鉴权失败
        """
        params = {"save_id": "1"}
        resp = requests.get(GET_SAVE_URL, params=params)
        print(f"[test_get_save_unauthorized] status={resp.status_code}, resp={resp.text}")
        assert resp.status_code in (401, 403)

# 测试流程说明：
# 1. test_get_save_success：先创建存档，再用返回的 save_id 获取，断言内容正确
# 2. test_get_save_not_found：用不存在的 save_id 获取，应返回404
# 3. test_get_save_missing_params：不传 save_id，应返回400
# 4. test_get_save_unauthorized：不带token，应鉴权失败
# 运行方法：pytest test_save_get.py
