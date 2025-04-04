Assisted Document Authoring (ADA) is a tool that helps you write Executable Documentation.

This is a proof of concept. It is not intended to be production code. It is hoped that this will provide inspiration and a reference implementation for similar tools, optimized for specific purposes.

ADA is designed to be used in conjunction with a standard text editor and uses an LLM to help create content. As a command-line tool it can easily be integrated with any editor, but at the time of writing it current assumes that VS Code is present on the system. This is easily changed though - patches welcome.

## Installation

ADA is written in Python. Ensure Python3 is available on your path and run the following commands to install dependencies:

```bash
pip install --upgrade pip setuptools
pip install .
```

Setup your environment, it's easiest to add this to a `.env` file in your workspace.

```bash
AZURE_OPENAI_API_KEY="<your key>"
AZURE_OPENAI_ENDPOINT="https://experimentalembeddedagent627202997579.cognitiveservices.azure.com/"
AZURE_OPENAI_MODEL_NAME="gpt-4o"
AZURE_OPENAI_DEPLOYMENT="gpt-4o"
AZURE_OPENAI_API_VERSION="2024-12-01-preview"
```

## Running

```bash
python app/main.py
```

`Document` is the top level object and represents the docuent you are authoring. It consists of three sections `Overview`, `Deployment` and `Summary`. By running the above command you are telling the system you want to create a `Document`. Each section is created in the same way and the user is given the opportunity to edit, or have the LLM edit, the generated section before moving on to the next. Each section becomes input to the generation of the next section.

When running the above command you will be asked for a title. Be thoughtful about the title. This will have a significant impact on the content generated. For example, "A Linux VM" is not a very good title, but "VM Based Web Application Hosting" provides significantly more information for the tool to work with.

### Section Guidance (optional)

Next you will be asked to provide any special instructions about the section currently being generated. This will be used as a prompt for the LLM. You can simply hit enter here and the LLM will use a generic prompt. However, you can use this prompt to guide the sections generation in a particular direction. For example, when generating the overview you might say "Use NGinx as the web server and provide both a public IP and JIT SSH access." When generating the "Deployment" you might say things like "Deploy into EastUS2 and use the prefix 'ADATestEnv_3_20' on all values that need to be world unique."

Once you press enter here the application will move on to the next step in the authoring process.

### Draft Content

Once the LLM has generated a draft for a section it will open it in the editor (currently requires VS Code, but adding support for others is easy). The author can then manually edit the content and/or provide an LLM prompt requesting changes. This can be repeated as many times as needed. Once the user is happy with the section they hit enter to proceed.

This process is repeated for each required section of the document.

### Final Document

Once all sections have been created and edited the final document will be opened in the editor.

## Future Ideas

- When generating the deployment section automatically run the generated doc and attempt to fix it. If it continues to fail then add the error into the doc and have the human look at it.
- Enable the ingestion of an existing Exec Doc for iterative editing
- Enable the ingestion of an existing (non-exec) doc and convert it to this format, then allow iterative editing (Merge with Naman's work)
- Make it a VS Code plugin
- Make it a CoPilot agent
- Generate steps individually rather than as a single whole
- Offer to generate suitable "next steps" documents once a document is complete

## Development Notes

Application code is in the `app` folder

`Document` is the top level document
`BaseObject` is the parent class for each of `Outline`, `Deployment`, `Summary` sections
`llm` is the interface to the LLM. To change the LLM provider change this implementation
`ui` is the UI interface. To provide a different UX create an alternative implementation

### Confgiuring VS Code

Assuming you have added the environment variables above into a `.env` file in the `ADA` root directory.

### Testing

There is a currently only very rudimentary tests. This are not exhaustive, but they are better than nothing. Each of the section classes have a simple test in their `__main__`.
