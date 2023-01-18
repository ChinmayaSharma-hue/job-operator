package v1
import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-
gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HelloType struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HelloSpec `json:"spec"`

	Status HelloTypeStatus `json:"status,omitempty"`
}

type HelloSpec struct {
	Message string `json:"message, omitempty"`
}

type HelloTypeStatus struct {
	Name string
}

// +k8s:deepcopy-
gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HelloTypeList struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ListMeta `json:"metadata,omitempty"`

	Items []HelloType `json:"items"
}