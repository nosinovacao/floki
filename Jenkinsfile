pipeline {
    agent { dockerfile true }
    stages {
        stage("go test"){
            steps{
                echo "====++++executing go test++++===="
                sh 'go test'
            }
            post{
                // always{
                //     echo "====++++always++++===="
                // }
                success{
                    echo "====++++go test executed succesfully++++===="
                }
                failure{
                    echo "====++++go test execution failed++++===="
                }
        
            }
        }
    }
}