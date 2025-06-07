from flask import Blueprint, request, jsonify
from flask_jwt_extended import create_access_token, jwt_required, get_jwt_identity
from extensions import db, bcrypt
from models import User, APIKey, ModelAPIKey
import uuid
import logging

api_key_routes = Blueprint('api_key', __name__)

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)

@api_key_routes.route('/generate_api_key', methods=['POST'])
@jwt_required()
def generate_api_key():
    logger.info("Generating API key")
    user_id = int(get_jwt_identity())  # Convert back to int
    data = request.get_json()
    logger.info(f"Generating API key for user {user_id} with data: {data}")
    real_api_keys = data.get('real_api_keys')
    project_name = data.get('project_name')

    if not real_api_keys or not isinstance(real_api_keys, list):
        return jsonify({"message": "A list of real API keys is required"}), 400
    
    if not project_name:
        return jsonify({"message": "Project name is required"}), 400
    

    key = str(uuid.uuid4().hex)
    api_key = APIKey(user_id=user_id, key=key, project_name=project_name)
    db.session.add(api_key)
    db.session.flush()

    for rk in real_api_keys:
        model_key = ModelAPIKey(api_key_id=api_key.id, real_api_key=rk['key'], llm_provider=rk['llm_provider'], model_name=rk['model_name'])
        db.session.add(model_key)

    db.session.commit()
    return jsonify({'api_key': key})


@api_key_routes.route('/api_keys', methods=['GET'])
@jwt_required()
def get_api_keys():
    user_id = int(get_jwt_identity())
    api_keys = APIKey.query.filter_by(user_id=user_id).all()
    logger.debug(User.query.filter_by(id=user_id).first().email)

    api_keys = [{"api_key": key.key, "project": key.project_name, 'created_at': key.created_at} for key in api_keys]


    if not api_keys:
        return jsonify({"message": "No API keys found"}), 404
    return jsonify([{
        'key': api_keys
    }])

@api_key_routes.route('/get_model_keys', methods=['GET'])
@jwt_required()
def get_model_keys():
    data = request.get_json()
    api_key = data.get('api_key')
    user_id = int(get_jwt_identity())
    if not api_key:
        return jsonify({"message": "API key is required"}), 400
    key_obj = APIKey.query.filter_by(key=api_key, user_id=user_id).first()
    if not key_obj:
        return jsonify({"message": "API key not found"}), 404
    record = ModelAPIKey.query.filter_by(api_key_id=key_obj.id).all()
    if not record:
        return jsonify({"message": "No model keys found for this API key"}), 404
    model_keys = [{"key": rk.real_api_key, "llm_provider": rk.llm_provider, "model_name": rk.model_name }for rk in record]
    return jsonify({"model_keys": model_keys})

@api_key_routes.route('/delete_api_key', methods=['DELETE'])
@jwt_required()
def delete_api_key():
    data = request.get_json()
    user_id = int(get_jwt_identity())
    api_key = data.get('api_key')
    if not api_key:
        return jsonify({"message": "API key is required"}), 400
    
    key_obj = APIKey.query.filter_by(key=api_key, user_id=user_id).first()
    if not key_obj:
        return jsonify({"message": "API key not found"}), 404
    
    db.session.delete(key_obj)
    db.session.commit()
    
    return jsonify({"message": "API key deleted successfully"})

@api_key_routes.route('/delete_model_key', methods=['DELETE'])
@jwt_required()
def delete_model_key():
    data = request.get_json()
    api_key = data.get('api_key')
    model_key = data.get('model_key')
    model_name = data.get('model_name')
    
    if not api_key or not model_key or not model_name:
        return jsonify({"message": "API key, model key, and model name are required"}), 400
    
    user_id = int(get_jwt_identity())
    key_obj = APIKey.query.filter_by(key=api_key, user_id=user_id).first()
    if not key_obj:
        return jsonify({"message": "API key not found"}), 404
    
    model_key_obj = ModelAPIKey.query.filter_by(api_key_id=key_obj.id, real_api_key=model_key, model_name=model_name).first()
    if not model_key_obj:
        return jsonify({"message": "Model key not found"}), 404
    
    db.session.delete(model_key_obj)
    db.session.commit()
    
    return jsonify({"message": "Model key deleted successfully"})

@api_key_routes.route('/add_model_key', methods=['PUT'])
@jwt_required()
def add_model_key():
    data = request.get_json()
    api_key = data.get('api_key')
    new_model_key = data.get('new_model_key')
    llm_provider = data.get('llm_provider')
    model_name = data.get('model_name')
    if not llm_provider or not model_name:
        return jsonify({"message": "LLM provider and model name are required"}), 400
    
    if not api_key or not new_model_key:
        return jsonify({"message": "API key and new model key are required"}), 400
    
    user_id = int(get_jwt_identity())
    key_obj = APIKey.query.filter_by(key=api_key, user_id=user_id).first()
    if not key_obj:
        return jsonify({"message": "API key not found"}), 404
    
    model_key = ModelAPIKey(api_key_id=key_obj.id, real_api_key=new_model_key, llm_provider=llm_provider, model_name=model_name)
    db.session.add(model_key)
    db.session.commit()
    
    return jsonify({"message": "Model key added successfully"})