from flask import Blueprint, request, jsonify
from flask_jwt_extended import create_access_token, jwt_required, get_jwt_identity
from extensions import db, bcrypt
from models import User, APIKey, ModelAPIKey
import uuid
import logging

auth_routes = Blueprint('auth', __name__)

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)

@auth_routes.route('/register', methods=['POST'])
def register():
    data = request.get_json()
    email = data.get('email')
    password = data.get('password')
    if User.query.filter_by(email=email).first():
        return jsonify({"message": "Email already exists"}), 400
    hashed = bcrypt.generate_password_hash(password).decode('utf-8')
    user = User(email=email, password_hash=hashed)
    db.session.add(user)
    db.session.commit()
    return jsonify({"message": "User registered successfully"})

@auth_routes.route('/login', methods=['POST'])
def login():
    data = request.get_json()
    email = data.get('email')
    user = User.query.filter_by(email=email).first()
    if user and bcrypt.check_password_hash(user.password_hash, data.get('password')):
        token = create_access_token(identity=str(user.id))  # Convert to string
        return jsonify({"access_token": token})
    return jsonify({"message": "Invalid credentials"}), 401



