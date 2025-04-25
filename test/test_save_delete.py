# -*- coding: utf-8 -*-
"""
测试 /api/save/delete 接口
"""
import pytest
import requests

DELETE_SAVE_URL = "http://localhost:8888/api/save/delete"
CREATE_SAVE_URL = "http://localhost:8888/api/save/create"
GET_SAVE_URL = "http://localhost:8888/api/save/get"

class TestSaveDelete:
    def create_save_and_get_id(self, token: str) -> str:
        payload = {
            "save_name": "Delete测试存档",
            "save_data": "内容",
            "save_type": "novel",
            "save_description": "章节1"
        }
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.post(CREATE_SAVE_URL, json=payload, headers=headers)
        assert resp.status_code == 200
        data = resp.json()
        assert "save_id" in data
        return data["save_id"]

    def test_delete_save_success(self, auto_register_and_login):
        """
        正常删除已存在存档
        """
        token = auto_register_and_login
        save_id = self.create_save_and_get_id(token)
        payload = {"save_id": save_id}
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.delete(DELETE_SAVE_URL, json=payload, headers=headers)
        print(f"[test_delete_save_success] status={resp.status_code}, resp={resp.text}")
        assert resp.status_code == 200
        data = resp.json()
        assert data["code"] == 200
        assert data["message"] == "删除成功"
        # 查询确认
        params = {"save_id": save_id}
        resp2 = requests.get(GET_SAVE_URL, params=params, headers=headers)
        assert resp2.status_code == 404

    def test_delete_save_not_found(self, auto_register_and_login):
        """
        删除不存在的存档，应返回404
        """
        token = auto_register_and_login
        payload = {"save_id": "not_exist_save_id"}
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.delete(DELETE_SAVE_URL, json=payload, headers=headers)
        print(f"[test_delete_save_not_found] status={resp.status_code}, resp={resp.text}")
        assert resp.status_code == 404
        data = resp.json()
        assert data["code"] == 404
        assert data["message"] == "保存项不存在"

    def test_delete_save_missing_params(self, auto_register_and_login):
        """
        缺少 save_id 参数，应返回400
        """
        token = auto_register_and_login
        payload = {}
        headers = {"Authorization": f"Bearer {token}"}
        resp = requests.delete(DELETE_SAVE_URL, json=payload, headers=headers)
        print(f"[test_delete_save_missing_params] status={resp.status_code}, resp={resp.text}")
        assert resp.status_code == 400
        data = resp.json()
        assert data["code"] == 400
        assert "缺少必需参数" in data["message"] or "请求参数不合法" in data["message"]

    def test_delete_save_unauthorized(self):
        """
        未授权请求，应返回401
        """
        payload = {"save_id": "xxx"}
        resp = requests.delete(DELETE_SAVE_URL, json=payload)
        print(f"[test_delete_save_unauthorized] status={resp.status_code}, resp={resp.text}")
        assert resp.status_code in (401, 403)
