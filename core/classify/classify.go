package main

import (
	"encoding/csv"
	"fmt"
	nb "github.com/pa-m/sklearn/naive_bayes"
	"github.com/pa-m/sklearn/pipeline"
	"github.com/pa-m/sklearn/preprocessing"
	"gonum.org/v1/gonum/mat"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {
	//inspired from https://scikit-learn.org/stable/auto_examples/preprocessing/plot_scaling_importance.html
	//results have a small diff from python ones because train_test_split is different
	//features, target := datasets.LoadWine().GetXY()
	//RandomState := uint64(42)
	//Xtrain, Xtest, Ytrain, Ytest := modelselection.TrainTestSplit(features, target, .30, RandomState)
	fileNameTrain := "F:\\研二\\毕业设计\\Network-Intrusion-Detection\\KDDCup 99\\classical\\binary\\kddtrain.csv"
	fsTrain, err := os.Open(fileNameTrain)
	if err != nil {
		log.Fatalf("can not open the file, err is %+v", err)
	}
	defer fsTrain.Close()

	fileNameTest := "F:\\研二\\毕业设计\\Network-Intrusion-Detection\\KDDCup 99\\classical\\binary\\kddtest.csv"
	fsTest, err := os.Open(fileNameTest)
	if err != nil {
		log.Fatalf("can not open the file, err is %+v", err)
	}
	defer fsTest.Close()

	r := csv.NewReader(fsTrain)

	protocolTypeDict := make(map[string]float64)
	serviceDict := make(map[string]float64)
	flagDict := make(map[string]float64)
	labelDict := make(map[string]float64)
	reLabelDict := make(map[float64]string)
	kddTrain := make([]float64, 0)
	kddTrainLabel := make([]float64, 0)

	protocolTypeCnt := 0
	serviceCnt := 0
	flagCnt := 0
	labelCnt := 0

	//rowLen := 0

	for {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			log.Fatalf("can not read, err is %+v", err)
		}
		if err == io.EOF {
			break
		}
		//rowLen = len(row)
		for i := range row {
			if (i >= 9 && i <= 21) || i == 42 {
				continue
			}

			if i == 1 {
				if v, ok := protocolTypeDict[row[i]]; ok {
					fmt.Print(fmt.Sprintf("%0.1f", v) + " ")
					kddTrain = append(kddTrain, v)
				} else {
					protocolTypeDict[row[i]] = float64(protocolTypeCnt)
					fmt.Print(fmt.Sprintf("%0.1f", protocolTypeDict[row[i]]) + " ")
					kddTrain = append(kddTrain, protocolTypeDict[row[i]])
					protocolTypeCnt++
				}
			} else if i == 2 {
				if v, ok := serviceDict[row[i]]; ok {
					fmt.Print(fmt.Sprintf("%0.1f", v) + " ")
					kddTrain = append(kddTrain, v)
				} else {
					serviceDict[row[i]] = float64(serviceCnt)
					fmt.Print(fmt.Sprintf("%0.1f", serviceDict[row[i]]) + " ")
					kddTrain = append(kddTrain, serviceDict[row[i]])
					serviceCnt++
				}
			} else if i == 3 {
				if v, ok := flagDict[row[i]]; ok {
					fmt.Print(fmt.Sprintf("%0.1f", v) + " ")
					kddTrain = append(kddTrain, v)
				} else {
					flagDict[row[i]] = float64(flagCnt)
					fmt.Print(fmt.Sprintf("%0.1f", flagDict[row[i]]) + " ")
					kddTrain = append(kddTrain, flagDict[row[i]])
					flagCnt++
				}
			} else if i == 41 {
				if v, ok := labelDict[row[i]]; ok {
					kddTrainLabel = append(kddTrainLabel, v)
				} else {
					labelDict[row[i]] = float64(labelCnt)
					reLabelDict[float64(labelCnt)] = row[i]
					kddTrainLabel = append(kddTrainLabel, flagDict[row[i]])
					labelCnt++
				}
			} else {
				v, _ := strconv.ParseFloat(row[i], 64)
				fmt.Print(row[i] + " ")
				kddTrain = append(kddTrain, v)
			}
		}
		fmt.Println()
	}

	//xxTrain := mat.NewDense(2,2, []float64{
	//	1,1,
	//	2,2,
	//})
	//yyTrain := mat.NewDense(2, 1, []float64{
	//	1,
	//	1,
	//})
	//xxTest := mat.NewDense(1,2,[]float64{
	//	1.5,1.5,
	//})
	//yyTest := mat.NewDense(1,1,[]float64{
	//	0,
	//})

	fmt.Println(len(kddTrain))
	features := mat.NewDense(125973, 28, kddTrain)

	target := mat.NewDense(125973, 1, kddTrainLabel)

	//RandomState := uint64(42)
	//xxTrain, xxTest, yyTrain, yyTest := modelselection.TrainTestSplit(features, target, .30, RandomState)

	std := preprocessing.NewStandardScaler()
	pca := preprocessing.NewPCA()
	pca.NComponents = 15
	gnb := nb.NewGaussianNB(nil, 1e-9)

	unscaledClf := pipeline.MakePipeline(pca, gnb)
	unscaledClf.Fit(features, target)
	//fmt.Printf("Prediction accuracy for the normal test dataset with PCA %.2f %%\n", 100*unscaledClf.Score(xxTest, yyTest))

	std = preprocessing.NewStandardScaler()
	pca = preprocessing.NewPCA()
	pca.NComponents = 15
	gnb = nb.NewGaussianNB(nil, 1e-9)
	clf := pipeline.MakePipeline(std, pca, gnb)
	clf.Fit(features, target)
	/*score := clf.Score(xxTest, yyTest)
	fmt.Printf("Prediction accuracy for the standardized test dataset with PCA %.2f %%\n", 100*score)*/

	xTest := mat.NewDense(1, 28, []float64{0, 0.0, 2.0, 1.0, 0, 0, 0,
		0, 0, 184, 25, 1, 1, 0, 0, 0.14, 0.06, 0, 255, 25, 0.1, 0.06, 0, 0, 1, 1, 0, 0})
	yPredit := mat.NewDense(1, 1, nil)
	clf.Predict(xTest, yPredit)

	fmt.Println(reLabelDict[yPredit.At(0, 0)])
	// Output:
	// Prediction accuracy for the normal test dataset with PCA 70.37 %
	// Prediction accuracy for the standardized test dataset with PCA 98.15 %

}
