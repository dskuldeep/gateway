from flask import Flask
import datetime
from extensions import db, bcrypt, jwt
from flask_migrate import Migrate

def create_app():
    app = Flask(__name__)
    app.config['SQLALCHEMY_DATABASE_URI'] = 'sqlite:///llm_gateway.db'
    app.config['JWT_SECRET_KEY'] = 'llm_gateway_secret'
    app.config['JWT_ACCESS_TOKEN_EXPIRES'] = datetime.timedelta(days=30)

    db.init_app(app)
    migrate = Migrate(app, db) 
    bcrypt.init_app(app)
    jwt.init_app(app)

    from routes import auth_routes, llm_routes, api_key_routes
    app.register_blueprint(auth_routes)
    app.register_blueprint(llm_routes)
    app.register_blueprint(api_key_routes)

    with app.app_context():
        from models import User, APIKey, ModelAPIKey
        db.create_all()

    return app

if __name__ == '__main__':
    app = create_app()
    app.run(debug=True)