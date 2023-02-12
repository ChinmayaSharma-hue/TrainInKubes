## Train In Kubes

Train In Kubes is a Kubernetes operator that streamlines the process of training a machine learning model. The operator is designed to monitor a custom resource called TrainInKube and trigger a series of operations when this resource is created.

The first step of the operator is to create a ConfigMap containing all the necessary parameters for training the model. This includes details such as the dataset location, model location, and the number of epochs to run the model for.

Next, the operator creates a job that builds the model using the user-defined Docker image. Once the model is built, another job is triggered that splits the data and stores it in the specified location.

The operator then performs data parallel mode of training by creating multiple jobs in a loop for each minibatch of data. In each iteration, the jobs perform feedforward and backpropagation on their respective portions of the data, calculate the gradients, and update the model weights by taking the average of the gradients.

The operator repeats this process until all minibatches have been passed through the model, and the next epoch is started. This process continues until all the specified number of epochs have been completed, resulting in a fully trained machine learning model.

Train In Kubes is a highly customizable operator that can be tailored to meet the specific needs of your machine learning project. With its streamlined and automated process, it makes training your machine learning model in Kubernetes a breeze.

The manifest for the TrainInKube custom resource:

```
apiVersion: trainink8s.com/v1alpha1
kind: TrainInKube
metadata:
  name: example-traininkube
spec:
  modelImage: <DOCKER_IMAGE_OF_MODEL>
  modelImagePullPolicy: "Never" or "IfNotPresent"
  epochs: <NUMBER_OF_EPOCHS>
  batchSize: <BATCH_SIZE>
  numberOfSamples: <NUMBER_OF_SAMPLES>
  preprocessedDatasetLocation: <PREPROCESSED_DATASET_LOCATION>
  splitDatasetLocation: <SPLIT_DATASET_LOCATION>
  modelsLocation: <MODEL_LOCATION>
```

### Potential Enhancements

- Currently, the operator creates new sets of jobs for each minibatch of data. This is not the most efficient way to perform data parallel training. A more efficient way would be to create a single job that performs the training on all the minibatches of data. This would reduce the number of jobs created and the amount of time it takes to train the model. Need to find a way to do this.
- The operator only supports a HostPath volume for storing the data (As it was easy to test the operator using HostPath volume). Need to find a way to support other types of volumes, including PersistentVolumes from a cloud provider.
- The operator only supports data parallel training. Need to find a way to support model parallel training.
- The operator could benefit from a web UI that allows users to create TrainInKube custom resources without having to write the manifest themselves.
- The operator could benefit from a web UI that allows users to monitor the progress of their training jobs.
- Need to find other ways to improve each of the jobs, such as the train job being able to load only the data it needs to train on, instead of loading the entire split dataset.
- Need to use helm to package the operator and make it easier to install.
- Faced some errors while trying to run the operator in another namespace other than the default namespace, with a role assigned to the service account. Need to find a way to fix this. 

### Installation Instructions
To install the Train In Kubes operator, follow the steps below:

1. Clone the repository containing the operator code to your local machine.

2. Navigate to the manifests folder.

3. Use the TrainInKube.yaml file to install the operator. This file also creates the necessary custom resource, service account, cluster role, and cluster role binding.

4. Apply the yaml file to your cluster by using the following command:

    ```
    kubectl apply -f TrainInKube.yaml
    ```

5. Wait for the operator to be deployed. You can check its status using the command:
    ```
    kubectl get pods
    ```
6. Verify the installation by checking if the custom resource, service account, cluster role, and cluster role binding have been created successfully in your cluster.

7. You can now create a TrainInKube object in your cluster to start training your model.

The examples directory contains the docker images for the jobs created by the operator. Can be used to test the operator.





