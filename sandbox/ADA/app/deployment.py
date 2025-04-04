import os
import settings
import subprocess
import ui
import datetime

from base_object import BaseObject

class Deployment(BaseObject):
    """
    Deployment class represents the deployment section of a document. It includes the necessary steps and configurations to deploy the documented workload in a specific environment.
    
    The deployment should be detailed and specific, providing all the necessary information to successfully deploy the workload.
    
    The deployment is the main part of the document. It is critical that the deployment is accurate and well-structured, as it sets the tone for the rest of the document.
    """
    passed_tests = False

    def __init__(self):
        super().__init__()

    def generate(self, document):
        """
        Generates the deployment section of the document.
        
        Args:
            document (Document): The document to which this deployment section belongs.
        """
        super().generate(document)
        self.test_and_fix_errors()

    def get_prompt(self, document):
        """
        Returns the user prompt for a documents deployment section.
        """
        content = ""
        for section in document.sections:
            if section.instance and section.instance.content:
                content += section.instance.content

        if not document.auto:
            prompt = ui.get_user_input(f"Provide any special instructions for the deployment section of the document titled {document.title}.\n")
            prompt += "\n\nCurrent content of the document:\n\n "
        else:
            prompt = "Write the deployment section for the document below:\n\n"

        prompt += content
        return prompt
    
    def test_and_fix_errors(self):
        """
        Test the Document in the Innovation Engine.
        If the test fails then another call to the LLM is made to try to fix the error.
        The resulting document is tested again.
        This cycle will be repeated until the test passes or the maximum number of attempts is reached.

        Args:
            max_tests (int): The maximum number of tests to run before giving up.
        returns:
            bool: True if the test passed, False otherwise. Note this can also be retrieved from `Deployment.passed_tests`.
        """
        self.error_log = []  # Initialize an error log to store error messages and timestamps
        i = 0
        while i < settings.MAX_TEST_RUNS and self.passed_tests == False:
            i += 1
            ui.info(f"Testing the generated deployment content in IE (attempt {i} of {settings.MAX_TEST_RUNS})")
            filename = self.filename()

            try:
                process = subprocess.Popen(f"ie test '{filename}.md'", shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                stdout, stderr = process.communicate()
                ui.info(stdout)
                if stderr:
                    ui.error(stderr)
                
                if process.returncode != 0:
                    # TODO: We need a better output on syserror - see https://github.com/Azure/InnovationEngine/issues/258
                    raise subprocess.CalledProcessError(process.returncode, process.args, output=stdout)
                
                self.passed_tests = True
            except subprocess.CalledProcessError as e:
                error_message = e.output
                timestamp = datetime.datetime.now().isoformat()
                self.error_log.append({"timestamp": timestamp, "error_message": error_message})  # Store error and timestamp
                
                if "ERROR: AADSTS70043: The refresh token has expired" in error_message:
                    ui.error("The refresh token has expired. Please re-authenticate with the command `az login --use-device-code`.")
                    break

                # TODO: Can we do better than this? Perhaps having a Copilot explain what the error means? Will that improve success rated? It will, at least, allow us to provide special instructions for common errors.
                ui.error(f"\nError executing the document: {error_message}")
                self.user_edits(self.__class__.__name__, f"Fix this error, thrown when executing the document.\n\n{error_message}")

        return self.passed_tests

if __name__ == "__main__":
    import ui
    from document import Document
    import datetime

    ui.title("Testing Deployment class...")
    
    ui.info("Testing automated creation of a deployment section of a document")
    document = Document(False, "Create a Resource Group", "Create a resource group in the UK South region")
    deployment = Deployment()
    document.auto = True # switch to auto so that the deployment will be created automatically
    ui.info("Create the Deployment section")
    deployment.generate(document)
    deployment.display()

    if deployment.passed_tests:
        ui.print_green("Successfully created unattended (automated) document.")
    else:
        ui.print_red("Failed to create unattended (automated) document.")
        exit(1)
        
    ui.info("Testing manual creation of a deployment section of a document")
    document = Document()
    document.title = "Create a basic Linux VM on Azure"
    
    deployment = Deployment()
    ui.info("Create the Deployment section")
    deployment.generate(document)
    deployment.display()

    ui.info("Saving Deployment to JSON file...")
    deployment.save("test.json")

    ui.info("Loading deployment from JSON file...")
    loaded_deployment = Deployment.load("test.json")

    assert deployment.meta_data == loaded_deployment.meta_data
    assert deployment.content == loaded_deployment.content

    ui.print_green("Loaded Deployment matches the original.")
    ui.info("Deleting test.json...")
    os.remove("test.json")
    
    deployment.test_and_fix_errors()

    print()
    deployment.delete()
    
    ui.info("Done.")