from google import genai
from google.genai import types
import json

def llm_router(prompt, models):
    client = genai.Client(
        api_key="AIzaSyBV9z9rxCcv6BimkH3-9jejdWfa8m71rRE",
        http_options=types.HttpOptions(api_version='v1alpha')
    )
    system_prompt = """
    <system prompt>
    Your job is to select the best model for the given prompt from the provided list of models.
    You will receive a prompt and a list of models. Your response should be a JSON object with the following structure:
    {
        "model_name": "selected_model_name",
        "llm_provider": "provider_name",
        "reasoning": "explanation for the choice"
    }
    The "model" field should contain the name of the model you select, and the "llm_provider" field should contain the name of the provider for that model.
    You should consider the capabilities of each model and select the one that is best suited for the prompt and explain with a short reasoning for the choice.
    If you cannot determine a suitable model, return an error message in the following format:
    {
        "error": "No suitable model found"
    }
    Do not include any additional text or explanations in your response.
    </system prompt>
    """
    prompt = f"{system_prompt}\n\n <prompt> {prompt} </prompt> \n\n <models> {models} </models>"
    # print(prompt)
    response = client.models.generate_content(
        model="gemini-2.0-flash", contents=prompt
    )
    cleaned_output = response.text.strip().strip("`").strip()
    if cleaned_output.startswith("json"):
        cleaned_output = cleaned_output[len("json"):].strip()
    return json.loads(cleaned_output)

# models = [
#         {
#             "key": "AIzaSyBV9z9rxCcv6BimkH3-9jejdWfa8m71rRE",
#             "llm_provider": "gemini",
#             "model_name": "gemini-2.0-flash"
#         },
#         {
#             "key": "AIzaSyD_xgqNgURJYLFBoqbHyXQMvRY5y5vh4uM",
#             "llm_provider": "anthropic",
#             "model_name": "claude-3-5-sonnet-20240620"
#         },
#         {
#             "key": "AIzaSyBV9z9rxCcv6BimkH3-9jejdWfa8m71rRE",
#             "llm_provider": "gemini",
#             "model_name": "gemini-2.5-flash-preview-05-20"
#         }
#     ]

# prompt = "Write a python script for fibionacci series generation"
# response = llm_router(prompt, json.dumps(models))
# print(response)  