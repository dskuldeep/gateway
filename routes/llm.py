from flask import Blueprint, request, jsonify
from models import APIKey, ModelAPIKey
import random
from modules import gemini
from modules import router

llm_routes = Blueprint('llm', __name__)

@llm_routes.route('/llm_query', methods=['POST'])
def llm_query():
    api_key = request.headers.get('X-API-Key')
    record = APIKey.query.filter_by(key=api_key).first()
    if not record:
        return jsonify({"message": "Invalid API key"}), 403
    
    model_keys = ModelAPIKey.query.filter_by(api_key_id=record.id).all()
    if not model_keys:
        return jsonify({"message": "No model keys configured"}), 500
    
    # selected_key = random.choice(model_keys)
    # selected_model = selected_key.model_name  # PLaceholder for model selection logic

    selected_model = router.llm_router(prompt=request.json.get('prompt'), models=str([{'model_name': mk.model_name, 'llm_provider': mk.llm_provider} for mk in model_keys]))

    
    if not selected_model:
        return jsonify({"message": "No model selected by the Router"}), 400

    response = None
    selected_key = ModelAPIKey.query.filter_by(api_key_id=record.id, model_name=selected_model['model_name'], llm_provider=selected_model['llm_provider']).first()
    print(f"Selected Key: {selected_key}")
    # Other LLM providers can be added here

    if selected_key.llm_provider == 'gemini':
        # Call the Gemini module to send the query
        prompt = request.json.get('prompt')
        response = gemini.send_query(prompt, selected_key.model_name, selected_key.real_api_key)

    if not response:
        return jsonify({"message": "Failed to get response from Model"}), 500

    return jsonify({'response':response, 'model_name': selected_key.model_name, 'llm_provider': selected_key.llm_provider, 'reasoning': selected_model.get('reasoning', 'No reasoning provided')})