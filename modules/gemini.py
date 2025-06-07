from google import genai
from google.genai import types

def send_query(prompt, model_name, api_key):
    client = genai.Client(
        api_key=api_key,
        http_options=types.HttpOptions(api_version='v1alpha')
    )
    response = client.models.generate_content(
        model=model_name, contents=prompt
    )
    return response.text