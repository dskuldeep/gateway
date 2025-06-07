from api import db

class User(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    email = db.Column(db.String(120), unique=True, nullable=False)
    password_hash = db.Column(db.String(128), nullable=False)
    created_at = db.Column(db.DateTime, nullable=False, default=db.func.current_timestamp())
    updated_at = db.Column(db.DateTime, nullable=False, default=db.func.current_timestamp(), onupdate=db.func.current_timestamp())
    api_keys = db.relationship('APIKey', backref='user', lazy=True, cascade='all, delete-orphan')


class APIKey(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    user_id = db.Column(db.Integer, db.ForeignKey('user.id'), nullable=False)
    key = db.Column(db.String(64), unique=True, nullable=False)
    project_name = db.Column(db.String(100), nullable=False)
    created_at = db.Column(db.DateTime, nullable=False, default=db.func.current_timestamp())
    updated_at = db.Column(db.DateTime, nullable=False, default=db.func.current_timestamp(), onupdate=db.func.current_timestamp())
    model_keys = db.relationship('ModelAPIKey', backref='api_key', lazy=True, cascade='all, delete-orphan')
    routing_logic = db.relationship('RoutingLogic', backref='api_key', lazy=True, uselist=False, cascade='all, delete-orphan')

class ModelAPIKey(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    api_key_id = db.Column(db.Integer, db.ForeignKey('api_key.id'), nullable=False)
    real_api_key = db.Column(db.String(128), nullable=False)
    llm_provider = db.Column(db.String(50), nullable=False)
    model_name = db.Column(db.String(100), nullable=False)
    created_at = db.Column(db.DateTime, nullable=False, default=db.func.current_timestamp())


class RoutingLogic(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    api_key_id = db.Column(db.Integer, db.ForeignKey('api_key.id'), nullable=False)
    routing_map = db.Column(db.JSON, nullable=False)
    routing_state = db.Column(db.String(50), nullable=False, default='active')
    created_at = db.Column(db.DateTime, nullable=False, default=db.func.current_timestamp())
    updated_at = db.Column(db.DateTime, nullable=False, default=db.func.current_timestamp(), onupdate=db.func.current_timestamp())